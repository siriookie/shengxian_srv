package handler

import (
	"awesomeProject/shengxian/mxshop_srv/userop_srv/global"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/model"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/proto/userfav"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (*UserOpServer) GetFavList(ctx context.Context, req *userfav.UserFavRequest) (*userfav.UserFavListResponse, error) {
	var rsp userfav.UserFavListResponse
	var userFavs []model.UserFav
	var userFavList  []*userfav.UserFavResponse
	//查询用户的收藏记录 不需要传goodsID
	//查询某件商品被哪些用户收藏了 只需要传goodID
	result := global.DB.Where(&model.UserFav{User:req.UserId, Goods:req.GoodsId}).Find(&userFavs)
	rsp.Total = int32(result.RowsAffected)

	for _, userFav := range userFavs {
		userFavList = append(userFavList, &userfav.UserFavResponse{
			UserId: userFav.User,
			GoodsId: userFav.Goods,
		})
	}

	rsp.Data = userFavList

	return &rsp, nil
}

func (*UserOpServer) AddUserFav(ctx context.Context, req *userfav.UserFavRequest) (*emptypb.Empty, error) {
	var userFav model.UserFav

	userFav.User = req.UserId
	userFav.Goods = req.GoodsId

	global.DB.Save(&userFav)

	return &emptypb.Empty{}, nil
}


func (*UserOpServer) DeleteUserFav(ctx context.Context, req *userfav.UserFavRequest) (*emptypb.Empty, error) {
	//用了Unscoped 代表采用的是物理删除 因为之前在建表的时候创建了联合唯一索引 如果不是物理删除的话 新建的时候会产生冲突
	if result := global.DB.Unscoped().Where("goods=? and user=?", req.GoodsId, req.UserId).Delete(&model.UserFav{}); result.RowsAffected == 0{
		return nil, status.Errorf(codes.NotFound, "收藏记录不存在")
	}
	return &emptypb.Empty{}, nil
}

func (*UserOpServer) GetUserFavDetail(ctx context.Context, req *userfav.UserFavRequest) (*emptypb.Empty, error) {
	var ufav model.UserFav
	if result := global.DB.Where("goods=? and user=?", req.GoodsId, req.UserId).Find(&ufav); result.RowsAffected == 0{
		return nil, status.Errorf(codes.NotFound, "收藏记录不存在")
	}
	return &emptypb.Empty{}, nil
}
