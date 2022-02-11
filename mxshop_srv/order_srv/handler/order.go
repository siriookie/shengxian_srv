package handler

import (
	"awesomeProject/shengxian/mxshop_srv/order_srv/global"
	"awesomeProject/shengxian/mxshop_srv/order_srv/model"
	goodspb "awesomeProject/shengxian/mxshop_srv/order_srv/proto/goods"
	invpb "awesomeProject/shengxian/mxshop_srv/order_srv/proto/inventory"
	orderpb "awesomeProject/shengxian/mxshop_srv/order_srv/proto/order"
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

type OrderService struct {
	orderpb.UnimplementedOrderServer
}

func OrderSnGen(userID int32) string {
	//订单号的生成规则
	/*
		年月日 时分秒 + 用户id + 两位随机数
	*/
	now := time.Now()
	rand.Seed(time.Now().UnixNano())
	orderSn := fmt.Sprintf("%d%d%d%d%d%d%d%d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Nanosecond(),
		userID, rand.Intn(90)+10) //如果直接写rand.intN(100)的话 0-10是一位数的，不满足需求，如果直接加10，90-100又会是三位数，所以取0-90
	return orderSn
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

//	CartItemList 获取用户的购物车列表
func (s *OrderService) CartItemList(c context.Context, req *orderpb.UserInfoRequest) (*orderpb.CartItemListResponse, error) {
	var shopCarts []model.ShoppingCart
	//去数据库查询
	result := global.DB.Where(&model.ShoppingCart{User: req.ID}).Find(&shopCarts)
	if result.Error != nil {
		return nil, result.Error
	}
	rsp := &orderpb.CartItemListResponse{
		Total: int32(result.RowsAffected),
	}
	for _, shopCart := range shopCarts {
		rsp.ShopCartInfoResponse = append(rsp.ShopCartInfoResponse, &orderpb.ShopCartInfoResponse{
			ID:      shopCart.ID,
			UserID:  shopCart.User,
			GoodsID: shopCart.Goods,
			Nums:    shopCart.Nums,
			Checked: *shopCart.Checked,
		})
	}
	return rsp, nil
}

//	CreateCartItem 往购物车里面添加商品
func (s *OrderService) CreateCartItem(c context.Context, req *orderpb.CartItemRequest) (*orderpb.ShopCartInfoResponse, error) {
	//1.添加一个新的  2.之前就添加过一次
	var shopCart model.ShoppingCart
	result := global.DB.Where(model.ShoppingCart{Goods: req.GoodsID, User: req.UserID}).First(&shopCart)
	if result.RowsAffected == 1 {
		shopCart.Nums += req.Nums
	} else {
		shopCart.User = req.UserID
		shopCart.Nums = req.Nums
		shopCart.Goods = req.GoodsID
		f := false
		shopCart.Checked = &f
	}
	global.DB.Save(&shopCart)
	rsp := &orderpb.ShopCartInfoResponse{ID: shopCart.ID}
	return rsp, nil
}

//	UpdateCartItem 更新购物车的选中状态和数量
func (s *OrderService) UpdateCartItem(c context.Context, req *orderpb.CartItemRequest) (*emptypb.Empty, error) {
	var shopCart model.ShoppingCart
	result := global.DB.First(&shopCart, req.ID)
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "记录不存在")
	}
	shopCart.Checked = &req.Checked
	if req.Nums > 0 {
		shopCart.Nums = req.Nums
	}
	global.DB.Save(&shopCart)
	return &emptypb.Empty{}, nil
}

func (s *OrderService) DeleteCartItem(c context.Context, req *orderpb.CartItemRequest) (*emptypb.Empty, error) {
	var shopCart model.ShoppingCart
	result := global.DB.Delete(&shopCart, req.ID)
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "记录不存在")
	}
	return &emptypb.Empty{}, nil
}

func (s *OrderService) GetOrderList(c context.Context, req *orderpb.OrderFilterRequest) (*orderpb.OrderListResponse, error) {
	var orders []model.OrderInfo
	var total int64
	global.DB.Where(&model.OrderInfo{User: req.UserID}).Count(&total)
	var rsp orderpb.OrderListResponse
	rsp.Total = int32(total)
	global.DB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Where(&model.OrderInfo{User: req.UserID}).Find(&orders)
	for _, order := range orders {
		rsp.Data = append(rsp.Data, &orderpb.OrderInfoResponse{
			ID:      order.ID,
			UserID:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Total:   order.OrderMount,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SignerMobile,
		})
	}
	return &rsp, nil
}

func (s *OrderService) GetOrderDetail(c context.Context, req *orderpb.OrderRequest) (*orderpb.OrderInfoDetailResponse, error) {
	var order model.OrderInfo
	var rsp orderpb.OrderInfoDetailResponse
	if result := global.DB.Where(&model.OrderInfo{BaseModel: model.BaseModel{ID: req.ID}, User: req.UserID}).First(&order); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	orderInfo := orderpb.OrderInfoResponse{}
	orderInfo.ID = order.ID
	orderInfo.UserID = order.User
	orderInfo.OrderSn = order.OrderSn
	orderInfo.PayType = order.PayType
	orderInfo.Status = order.Status
	orderInfo.Post = order.Post
	orderInfo.Total = order.OrderMount
	orderInfo.Address = order.Address
	orderInfo.Name = order.SignerName
	orderInfo.Mobile = order.SignerMobile

	rsp.OrderInfo = &orderInfo
	var orderGoods []model.OrderGoods
	if result := global.DB.Where(&model.OrderGoods{Order: order.ID}).Find(&orderGoods); result.Error != nil {
		return nil, result.Error
	}
	for _, orderGood := range orderGoods {
		rsp.Goos = append(rsp.Goos, &orderpb.OrderItemResponse{
			ID:         orderGood.Goods,
			GoodsName:  orderGood.GoodsName,
			GoodsPrice: orderGood.GoodsPrice,
			GoodsImage: orderGood.GoodsImage,
			Nums:       orderGood.Nums,
		})
	}
	return &rsp, nil
}

type OrderListener struct {
	Code        codes.Code
	Detail      string
	ID          int32
	OrderAmount float32
}

//  When send transactional prepare(half) message succeed, this method will be invoked to execute local transaction.
func (t *OrderListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	checked := true
	var orderInfo model.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)
	var goodsIds []int32
	var shopCarts []model.ShoppingCart
	goodsNumMap := make(map[int32]int32)
	result := global.DB.Where(&model.ShoppingCart{User: orderInfo.User, Checked: &checked}).Find(&shopCarts)
	if result.RowsAffected == 0 {
		t.Code = codes.InvalidArgument
		t.Detail = "您的购物车里面还没有选中要结算的商品"
		//购物车中没有商品 而CreateOrder中创建了一个新的订单号，发送了一个半消息到rocketmq，现在商品都没有自然也不用建订单了，直接对半消息进行回滚
		return primitive.RollbackMessageState
	}
	for _, shopCart := range shopCarts {
		goodsIds = append(goodsIds, shopCart.Goods)
		goodsNumMap[shopCart.Goods] = shopCart.Nums
	}
	//跨服务调用
	goods, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &goodspb.BatchGoodsIdInfo{Id: goodsIds})
	if err != nil {
		t.Code = codes.Internal
		t.Detail = "批量查询商品信息失败"
		return primitive.RollbackMessageState
	}
	var orderAmount float32
	var orderGoods []*model.OrderGoods
	var goodsInvInfo []*invpb.GoodsInvInfo
	for _, good := range goods.Data {
		orderAmount += good.ShopPrice * float32(goodsNumMap[good.Id])
		orderGoods = append(orderGoods, &model.OrderGoods{
			Goods:      good.Id,
			GoodsName:  good.Name,
			GoodsImage: good.GoodsFrontImage,
			GoodsPrice: good.ShopPrice,
			Nums:       goodsNumMap[good.Id],
		})
		goodsInvInfo = append(goodsInvInfo, &invpb.GoodsInvInfo{
			GoodsID: good.Id,
			Num:     goodsNumMap[good.Id],
		})
	}

	//库存扣减微服务调用,有可能出现因为网络原因造成的没有拿到返回值，需要对sell返回的所有状态码进行判断，如果不是自己定义的状态码肯定就是网络状态出了问题
	_, err = global.InventorySrvClient.Sell(context.Background(), &invpb.SellInfo{
		GoodsInfo: goodsInvInfo,
		OrderSn: orderInfo.OrderSn,
	})
	if err != nil {
		t.Code = codes.ResourceExhausted
		t.Detail = "库存不足"
		return primitive.RollbackMessageState
	}
	//生成订单
	tx := global.DB.Begin()
	orderInfo.OrderMount = orderAmount

	if res := tx.Save(&orderInfo); res.RowsAffected == 0 {
		tx.Rollback()
		t.Code = codes.Internal
		t.Detail = "创建订单失败"
		//订单创建失败了 但是库存已经进行了扣减 所以要对已经扣减了的库存进行归还操作
		return primitive.CommitMessageState
	}
	t.OrderAmount = orderAmount
	t.ID = orderInfo.ID
	//把订单商品表的外键加上
	for _, orderGood := range orderGoods {
		orderGood.Order = orderInfo.ID
	}
	//批量插入orderGoods
	if res := tx.CreateInBatches(orderGoods, 100); res.RowsAffected == 0 {
		t.Code = codes.Internal
		t.Detail = "创建订单失败"
		tx.Rollback()
		return primitive.CommitMessageState

	}

	//删除购物车里的记录
	if result := tx.Where(&model.ShoppingCart{User: orderInfo.User, Checked: &checked}).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		tx.Rollback()
		t.Code = codes.Internal
		t.Detail = "删除购物车记录失败"
		return primitive.CommitMessageState

	}
	//发送取消订单的延迟消息
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{"127.0.0.1:9876"}),//等价下面的代码,下面的方法有点老
		//producer.WithNsResolver(primitive.NewPassthroughResolver([]string{"127.0.0.1:9876"})),
		producer.WithRetry(2),
	)
	if err != nil{
		panic(err)
	}
	if err = p.Start();err != nil{panic(err)}	//把producer启动起来
	msg1 := primitive.NewMessage("order_timeout",msg.Body)
	msg1.WithDelayTimeLevel(16) //延迟发送消息，对应级别有：1s 5s 10s 30s 1m 2m 3m 4m 5m 6m 7m 8m 9m 10m 20m 30m 1h 2h
	res,err := p.SendSync(context.Background(),msg1)
	if err != nil{
		zap.S().Error("发送消息失败:%s",err)
		tx.Rollback()
		t.Code = codes.Internal
		t.Detail = "发送延时消息失败"
		return primitive.CommitMessageState
	}
	fmt.Println(res.String())
	if err = p.Shutdown();err != nil{zap.S().Error(err)}

	//提交事务
	tx.Commit()
	return primitive.RollbackMessageState //代表本地执行的逻辑没有问题 可以把半消息发出去了
}

// When no response to prepare(half) message. broker will send check message to check the transaction status, and this
// method will be invoked to get local transaction status.
func (t *OrderListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	var orderInfo model.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)
	//拿到了消息队列里面没有处理到的消息
	if res := global.DB.Where(model.OrderInfo{OrderSn: orderInfo.OrderSn}).First(&orderInfo);res.RowsAffected == 0{
		return primitive.CommitMessageState	//并不一定能证明代码语句真的运行到了sell的地方
	}
	return primitive.RollbackMessageState //代表回查发现有已经生成了订单，可以将归还库存的消息给撤了

}

func (s *OrderService) CreateOrder(c context.Context, req *orderpb.OrderRequest) (*orderpb.OrderInfoResponse, error) {
	//新建一个rocketmq的半消息
	orderListener := OrderListener{}
	p, err := rocketmq.NewTransactionProducer(
		&orderListener,
		producer.WithNameServer([]string{"127.0.0.1:9876"}), //等价下面的代码,下面的方法有点老
		//producer.WithNsResolver(primitive.NewPassthroughResolver([]string{"127.0.0.1:9876"})),
		producer.WithRetry(2),
	)
	if err != nil {
		zap.S().Error(err)
		return nil, err
	}
	if err = p.Start(); err != nil {
		zap.S().Error(err)
		return nil, err
	}
	order := model.OrderInfo{
		OrderSn:      OrderSnGen(req.UserID),
		Address:      req.Address,
		SignerName:   req.Name,
		SignerMobile: req.Mobile,
		Post:         req.Post,
		User:         req.UserID,
	}
	jsonStr, _ := json.Marshal(order)

	res, err := p.SendMessageInTransaction(context.Background(), primitive.NewMessage("order_reback", jsonStr))
	if err != nil {
		zap.S().Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if res.State == primitive.CommitMessageState {
		return nil, status.Error(codes.Internal, "新建订单失败")
	}
	fmt.Println(res.String())
	p.Shutdown()
	return &orderpb.OrderInfoResponse{ID: orderListener.ID,OrderSn: order.OrderSn,Total: orderListener.OrderAmount},nil

}

func (s *OrderService) UpdateOrderStatus(c context.Context, req *orderpb.OrderStatus) (*emptypb.Empty, error) {
	result := global.DB.Model(&model.OrderInfo{}).Where("order_sn = ?", req.OrderSn).Update("status", req.Status)
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "订单不存在")
	}
	return &emptypb.Empty{}, nil
}


func OrderTimeout(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error){
	for i := range ext {
		var orderInfo model.OrderInfo
		_ = json.Unmarshal(ext[i].Body, &orderInfo)
		//查询订单的支付状态,如果没有支付就要归还库存
		var order model.OrderInfo
		tx := global.DB.Begin()
		if res := tx.Model(model.OrderInfo{}).Where(model.OrderInfo{OrderSn: orderInfo.OrderSn}).First(&order); res.RowsAffected == 0 {
			return consumer.ConsumeSuccess, nil
		}
		if order.Status != "TRADE_SUCCESS" {
			//归还库存,可以往order_reback中发送一个消息
			p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"127.0.0.1:9876"}))
			if err != nil {
				zap.S().Error("生成producer失败")
				return 0, err
			}
			if err = p.Start(); err != nil {
				zap.S().Error("生成producer失败")
				return 0, err
			}
			_, _ = p.SendSync(context.Background(), primitive.NewMessage("order_reback", ext[i].Body))
			if err = p.Shutdown(); err != nil {
				return 0, err
			}
			//修改订单状态为未支付
			order.Status = "TRADE_CLOSED"
			tx.Save(&order)
			tx.Commit()
		}
	}
	return consumer.ConsumeSuccess,nil
}