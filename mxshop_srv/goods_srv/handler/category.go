package handler

import (
	"awesomeProject/shengxian/mxshop_srv/goods_srv/global"
	"awesomeProject/shengxian/mxshop_srv/goods_srv/model"
	goodspb "awesomeProject/shengxian/mxshop_srv/goods_srv/proto/goods"
	"context"
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

////商品分类
func (s *GoodsService) GetAllCategorysList(c context.Context, req *emptypb.Empty) (*goodspb.CategoryListResponse, error) {
	var categorys []model.Category
	global.DB.Where(&model.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categorys)
	b, _ := json.Marshal(&categorys)
	return &goodspb.CategoryListResponse{
		JsonData: string(b),
	}, nil
}

////获取子分类
func (s *GoodsService) GetSubCategory(c context.Context, req *goodspb.CategoryListRequest) (*goodspb.SubCategoryListResponse, error){
	categoryListRsp := goodspb.SubCategoryListResponse{}
	var category model.Category
	if result := global.DB.First(&category,req.Id);result.RowsAffected == 0{
		return nil,status.Error(codes.NotFound,"商品分类不存在")
	}
	categoryListRsp.Info = &goodspb.CategoryInfoResponse{
		Id: category.ID,
		Name: category.Name,
		Level: category.Level,
		IsTab: category.IsTab,
		ParentCategory: category.ParentCategoryID,
	}

	var subCategorys []model.Category
	var subCategorysRsp []*goodspb.CategoryInfoResponse
	preloads := "SubCategory"
	if category.Level == 1{
		preloads = "SubCategory.SubCategory"
	}
	global.DB.Where(&model.Category{ParentCategoryID: req.Id}).Preload(preloads).Find(&subCategorys)
	for _,subCategory := range subCategorys{
		subCategorysRsp = append(subCategorysRsp,&goodspb.CategoryInfoResponse{
			Id: subCategory.ID,
			Name: subCategory.Name,
			Level: subCategory.Level,
			IsTab: subCategory.IsTab,
			ParentCategory: subCategory.ParentCategoryID,
		})
	}

	categoryListRsp.SubCategorys = subCategorysRsp
	return &categoryListRsp,nil
}

func (s *GoodsService) CreateCategory(c context.Context, req *goodspb.CategoryInfoRequest) (*goodspb.CategoryInfoResponse, error){
	category := model.Category{}
	cMap := map[string]interface{}{}
	cMap["name"] = req.Name
	cMap["level"] = req.Level
	cMap["is_tab"] = req.IsTab
	if req.Level != 1{
		//去查询父类目是否存在
		var parent model.Category
		parentCategory := global.DB.Where("parent_category_id=?",req.ParentCategory).First(&parent)
		if parentCategory.RowsAffected == 0{
			return nil,status.Error(codes.InvalidArgument,"不存在父分类")
		}
		if req.Level > parent.Level{
			return nil,status.Error(codes.InvalidArgument,"父分类的level必须大于子分类")
		}
		cMap["parent_category_id"] = req.ParentCategory
	}
	_ = global.DB.Model(&model.Category{}).Create(cMap)
	return &goodspb.CategoryInfoResponse{Id:int32(category.ID)}, nil

}
func (s *GoodsService) DeleteCategory(c context.Context, req *goodspb.DeleteCategoryRequest) (*emptypb.Empty, error){
	if result := global.DB.Delete(&model.Category{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	return &emptypb.Empty{}, nil
}
func (s *GoodsService) UpdateCategory(c context.Context, req *goodspb.CategoryInfoRequest) (*emptypb.Empty, error){
	var category model.Category
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.ParentCategory != 0 {	//proto传输的默认值是0
		category.ParentCategoryID = req.ParentCategory
	}
	if req.Level != 0 {
		category.Level = req.Level
	}
	if req.IsTab {
		category.IsTab = req.IsTab
	}
	global.DB.Save(&category)

	return &emptypb.Empty{}, nil
}
