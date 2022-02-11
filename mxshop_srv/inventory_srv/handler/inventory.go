package handler

import (
	"awesomeProject/shengxian/mxshop_srv/inventory_srv/global"
	"awesomeProject/shengxian/mxshop_srv/inventory_srv/model"
	invpb "awesomeProject/shengxian/mxshop_srv/inventory_srv/proto/inventory"
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/go-redsync/redsync/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type InventoryService struct {
	invpb.UnimplementedInventoryServer
}

func (s *InventoryService) SetInv(c context.Context, req *invpb.GoodsInvInfo) (*emptypb.Empty, error) {
	var inv model.Inventory
	global.DB.Where("goods = ?", req.GoodsID).First(&inv)
	if inv.Goods == 0 { //没有查询到的情况
		inv.Goods = req.GoodsID
	}
	inv.Stocks = req.Num
	global.DB.Save(&inv)
	return &emptypb.Empty{}, nil
}

func (s *InventoryService) InvDetail(c context.Context, req *invpb.GoodsInvInfo) (*invpb.GoodsInvInfo, error) {
	var inv model.Inventory
	if res := global.DB.Where("goods = ?", req.GoodsID).First(&inv); res.RowsAffected == 0 {
		return nil, status.Error(codes.InvalidArgument, "不存在的库存信息")
	}
	return &invpb.GoodsInvInfo{
		GoodsID: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

func (s *InventoryService) Sell(c context.Context, req *invpb.SellInfo) (*emptypb.Empty, error) {
	tx := global.DB.Begin()             //开始数据库事务
	rs := redsync.New(global.RedisPool) //在池里面拿到一个redsync
	sellDetail := model.StockSellDetail{
		OrderSn: req.OrderSn,
		Status:  1, //默认表示为已经扣减了
	}
	var details []model.GoodsDetail
	for _, good := range req.GoodsInfo {
		details = append(details, model.GoodsDetail{
			Goods: good.GoodsID,
			Num:   good.Num,
		})
		var inv model.Inventory
		// 通过设置相同的锁名来对每个协程加同一把锁
		mutexName := fmt.Sprintf("goods:inv:num:%d", good.GoodsID)
		mutex := rs.NewMutex(mutexName)
		if err := mutex.Lock(); err != nil {
			zap.S().Errorf("获取redis分布式锁异常%s", err.Error())
			return nil, status.Error(codes.Internal, err.Error())
		}
		zap.S().Debug("拿到了一把锁")
		if res := global.DB.Where("goods = ?", good.GoodsID).First(&inv); res.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			if _, err := mutex.Unlock(); err != nil {
				zap.S().Errorf("释放redis分布式锁异常%s", err.Error())
			}
			return nil, status.Error(codes.InvalidArgument, "不存在的库存信息")
		}
		if inv.Stocks < good.Num {
			tx.Rollback() //回滚之前的操作
			if _, err := mutex.Unlock(); err != nil {
				zap.S().Errorf("释放redis分布式锁异常%s", err.Error())
			}
			return nil, status.Error(codes.ResourceExhausted, "库存不足")
		}
		inv.Stocks -= good.Num
		tx.Save(&inv)
		if _, err := mutex.Unlock(); err != nil {
			zap.S().Errorf("释放redis分布式锁异常%s", err.Error())
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	sellDetail.Detail = details
	//写selldetail表
	if res := tx.Create(&sellDetail); res.RowsAffected == 0 {
		//证明没写成功
		tx.Rollback()
		return nil, status.Error(codes.Internal, "保存库存扣减表失败")
	}
	tx.Commit() //手动提交操作
	return &emptypb.Empty{}, nil
}

func (s *InventoryService) ReBack(c context.Context, req *invpb.SellInfo) (*emptypb.Empty, error) {
	//库存归还：1：订单超时的归还 2.订单创建失败 3.用户取消订单
	tx := global.DB.Begin()
	for _, good := range req.GoodsInfo {
		var inv model.Inventory
		if res := global.DB.Where("goods = ?", good.GoodsID).First(&inv); res.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Error(codes.InvalidArgument, "不存在的库存信息")
		}
		inv.Stocks += good.Num
		tx.Save(&inv)
	}
	tx.Commit() //手动提交操作
	return &emptypb.Empty{}, nil
}

func AutoReback(ctx context.Context, msg ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	type OrderInfo struct {
		orderSn string
	}
	for i := range msg {
		//归还库存 需要知道要归还的订单信息，但是要想清楚重复归还的问题
		//所以这个接口应该确保幂等性，不能因为消息的重复发送导致订单库存归还多次，没有扣减的库存不要进行归还
		//解决办法 新建一张表 表里面记录订单的扣减细节和归还细节
		var orderInfo OrderInfo
		_ = json.Unmarshal(msg[i].Body, &orderInfo)
		//将库存加回去，将sellDetail的status设置为2，要创建本地事务
		tx := global.DB.Begin()
		var sellDetail model.StockSellDetail
		if res := tx.Where(&model.StockSellDetail{OrderSn: orderInfo.orderSn,Status: 1}).First(&sellDetail);res.RowsAffected==0{
			return consumer.ConsumeSuccess,nil
		}
		//如果查询到，那么逐个归还库存
		for _,orderGood := range sellDetail.Detail{
			//这个语句在mysql中会变成update xx set stock=stock+？ mysql会锁住这条记录来保证并发  如果有索引就会是行锁 如果没有索引就是表锁 优点：简单 缺点：性能差
			if res := tx.Model(&model.Inventory{}).Where(&model.Inventory{Goods: orderGood.Goods}).Update("stocks",gorm.Expr("stocks+?",orderGood.Num));res.RowsAffected==0{
				tx.Rollback()
				return consumer.ConsumeRetryLater,nil
			}
		}
		sellDetail.Status = 2
		if res := tx.Model(&model.StockSellDetail{}).Where(&model.StockSellDetail{OrderSn: orderInfo.orderSn}).Update("status",2);res.RowsAffected==0{
			tx.Rollback()
			return consumer.ConsumeRetryLater,nil
		}
		tx.Commit()
		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess,nil
}
