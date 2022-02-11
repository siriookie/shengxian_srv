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

////品牌分类
func (s *GoodsService) CategoryBrandList(c context.Context,req *goodspb.CategoryBrandFilterRequest) (*goodspb.CategoryBrandListResponse, error){
	var categoryBrands []model.GoodsCategoryBrand
	categoryBrandListResponse := goodspb.CategoryBrandListResponse{}

	var total int64
	//获取总数
	global.DB.Model(&model.GoodsCategoryBrand{}).Count(&total)
	categoryBrandListResponse.Total = int32(total)
	global.DB.Preload("Category").Preload("Brands").Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&categoryBrands)

	var categoryResponses []*goodspb.CategoryBrandResponse
	for _, categoryBrand := range categoryBrands {
		categoryResponses = append(categoryResponses, &goodspb.CategoryBrandResponse{
			Id: categoryBrand.ID,
			Category: &goodspb.CategoryInfoResponse{
				Id: categoryBrand.Category.ID,
				Name: categoryBrand.Category.Name,
				Level: categoryBrand.Category.Level,
				IsTab: categoryBrand.Category.IsTab,
				ParentCategory: categoryBrand.Category.ParentCategoryID,
			},
			Brand: &goodspb.BrandInfoResponse{
				Id:   categoryBrand.Brands.ID,
				Name: categoryBrand.Brands.Name,
				Logo: categoryBrand.Brands.Logo,
			},
		})
	}
	categoryBrandListResponse.Data = categoryResponses
	return &categoryBrandListResponse,nil
}
////通过category获取brands
func (s *GoodsService) GetCategoryBrandList(c context.Context, req *goodspb.CategoryInfoRequest) (*goodspb.BrandListResponse, error){
	brandListResponse := goodspb.BrandListResponse{}
	var category model.Category
	if result := global.DB.Find(&category, req.Id).First(&category); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "商品分类不存在")
	}
	var categoryBrands []model.GoodsCategoryBrand
	if result := global.DB.Preload("Brands").Where(&model.GoodsCategoryBrand{CategoryID: req.Id}).Find(&categoryBrands); result.RowsAffected > 0 {
		brandListResponse.Total = int32(result.RowsAffected)
	}
	var brandInfoResponses []*goodspb.BrandInfoResponse
	for _, categoryBrand := range categoryBrands {
		brandInfoResponses = append(brandInfoResponses, &goodspb.BrandInfoResponse{
			Id: categoryBrand.Brands.ID,
			Name: categoryBrand.Brands.Name,
			Logo: categoryBrand.Brands.Logo,
		})
	}
	brandListResponse.Data = brandInfoResponses

	return &brandListResponse, nil
}
func (s *GoodsService) CreateCategoryBrand(c context.Context, req *goodspb.CategoryBrandRequest) (*goodspb.CategoryBrandResponse, error){
	var category model.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "商品分类不存在")
	}

	var brand model.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}
	categoryBrand := model.GoodsCategoryBrand{
		CategoryID: req.CategoryId,
		BrandsID: req.BrandId,
	}
	global.DB.Save(&categoryBrand)
	return &goodspb.CategoryBrandResponse{Id: categoryBrand.ID}, nil

}
func (s *GoodsService) DeleteCategoryBrand(c context.Context, req *goodspb.CategoryBrandRequest) (*emptypb.Empty, error){
	if result := global.DB.Delete(&model.GoodsCategoryBrand{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌分类不存在")
	}
	return &emptypb.Empty{}, nil
}
func (s *GoodsService) UpdateCategoryBrand(c context.Context, req *goodspb.CategoryBrandRequest) (*emptypb.Empty, error){
	var categoryBrand model.GoodsCategoryBrand
	if result := global.DB.First(&categoryBrand, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌分类不存在")
	}
	var category model.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "商品分类不存在")
	}
	var brand model.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}
	categoryBrand.CategoryID = req.CategoryId
	categoryBrand.BrandsID = req.BrandId

	global.DB.Save(&categoryBrand)

	return &emptypb.Empty{}, nil
}
