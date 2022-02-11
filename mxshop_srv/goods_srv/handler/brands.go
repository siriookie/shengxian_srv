package handler

import (
	"awesomeProject/shengxian/mxshop_srv/goods_srv/global"
	"awesomeProject/shengxian/mxshop_srv/goods_srv/model"
	goodspb "awesomeProject/shengxian/mxshop_srv/goods_srv/proto/goods"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

////品牌
func (s *GoodsService) BrandList(c context.Context, req *goodspb.BrandFilterRequest) (*goodspb.BrandListResponse, error){
	brandListRsp := goodspb.BrandListResponse{}
	var brands []model.Brands
	//分页查询
	result := global.DB.Scopes(Paginate(int(req.Pages),int(req.PagePerNums))).Find(&brands)
	if result.Error != nil{
		return nil,result.Error
	}
	//查询总数
	var total int64
	global.DB.Model(&model.Brands{}).Count(&total)


	var brandResponses []*goodspb.BrandInfoResponse
	for _,brand := range brands{
		brandResponse := goodspb.BrandInfoResponse{
			Id: brand.ID,
			Name:brand.Name,
			Logo: brand.Logo,
		}
		brandResponses = append(brandResponses,&brandResponse)
	}
	brandListRsp.Data = brandResponses
	brandListRsp.Total = int32(total)
	return &brandListRsp,nil
}
func (s *GoodsService) CreateBrand(c context.Context, req *goodspb.BrandRequest) (*goodspb.BrandInfoResponse, error){
	result := global.DB.First(&model.Brands{},req.Id)
	if result.RowsAffected > 0 {
		return nil,status.Error(codes.InvalidArgument,"品牌已存在")
	}
	brand := model.Brands{}
	brand.Name = req.Name
	brand.Logo = req.Logo
	global.DB.Save(&brand)
	return &goodspb.BrandInfoResponse{
		Id: brand.ID,
	},nil
}
func (s *GoodsService) DeleteBrand(c context.Context, req *goodspb.BrandRequest) (*emptypb.Empty, error){
	if result := global.DB.Delete(&model.Brands{},req.Id);result.RowsAffected == 0{
		return nil,status.Error(codes.NotFound,"品牌不存在")
	}
	return &emptypb.Empty{},nil
}

func (s *GoodsService)UpdateBrand(c context.Context, req *goodspb.BrandRequest) (*emptypb.Empty, error){
	var brand model.Brands
	if result := global.DB.First(&brand,req.Id);result.RowsAffected == 0{
		return nil,status.Error(codes.NotFound,"品牌不存在")
	}
	if req.Name != ""{
		brand.Name = req.Name
	}
	if req.Logo != ""{
		brand.Logo = req.Logo
	}
	global.DB.Save(&brand)
	return &emptypb.Empty{},nil
}