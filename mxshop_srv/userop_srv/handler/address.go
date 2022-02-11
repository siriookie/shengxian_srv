package handler

import (
	"awesomeProject/shengxian/mxshop_srv/userop_srv/global"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/model"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/proto/addr"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)


func (*UserOpServer) GetAddressList(ctx context.Context, req *addr.AddressRequest) (*addr.AddressListResponse, error) {
	var addresses []model.Address
	var rsp addr.AddressListResponse
	var addressResponse []*addr.AddressResponse

	if result := global.DB.Where(&model.Address{User:req.UserId}).Find(&addresses); result.RowsAffected != 0 {
		rsp.Total = int32(result.RowsAffected)
	}

	for _, address := range addresses {
		addressResponse = append(addressResponse, &addr.AddressResponse{
			Id: address.ID,
			UserId: address.User,
			Province: address.Province,
			City: address.City,
			District: address.District,
			Address:  address.Address,
			SignerName: address.SignerName,
			SignerMobile:address.SignerMobile,
		})
	}
	rsp.Data = addressResponse

	return &rsp, nil
}

func (*UserOpServer) CreateAddress(ctx context.Context, req *addr.AddressRequest) (*addr.AddressResponse, error) {
	var address model.Address

	address.User = req.UserId
	address.Province = req.Province
	address.City = req.City
	address.District = req.District
	address.Address = req.Address
	address.SignerName = req.SignerName
	address.SignerMobile = req.SignerMobile

	global.DB.Save(&address)

	return &addr.AddressResponse{Id:address.ID}, nil
}

func (*UserOpServer) DeleteAddress(ctx context.Context, req *addr.AddressRequest) (*emptypb.Empty, error) {
	if result := global.DB.Where("id=? and user=?", req.Id, req.UserId).Delete(&model.Address{}); result.RowsAffected == 0{
		return nil, status.Errorf(codes.NotFound, "收货地址不存在")
	}
	return &emptypb.Empty{}, nil
}

func (*UserOpServer) UpdateAddress(ctx context.Context, req *addr.AddressRequest) (*emptypb.Empty, error) {
	var address model.Address

	if result := global.DB.Where("id=? and user=?", req.Id, req.UserId).First(&address); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}

	if address.Province != "" {
		address.Province = req.Province
	}

	if address.City != "" {
		address.City = req.City
	}

	if address.District != "" {
		address.District = req.District
	}

	if address.Address != "" {
		address.Address = req.Address
	}

	if address.SignerName != "" {
		address.SignerName = req.SignerName
	}

	if address.SignerMobile != "" {
		address.SignerMobile = req.SignerMobile
	}

	global.DB.Save(&address)

	return &emptypb.Empty{}, nil
}