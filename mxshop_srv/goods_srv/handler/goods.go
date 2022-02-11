package handler

import (
	"awesomeProject/shengxian/mxshop_srv/goods_srv/global"
	"awesomeProject/shengxian/mxshop_srv/goods_srv/model"
	goodspb "awesomeProject/shengxian/mxshop_srv/goods_srv/proto/goods"
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type GoodsService struct {
	goodspb.UnimplementedGoodsServer
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

func Model2Rsp(goods model.Goods) goodspb.GoodsInfoResponse {
	goodsInfoRsp := goodspb.GoodsInfoResponse{
		Id:              goods.ID,
		CategoryId:      goods.CategoryID,
		Name:            goods.Name,
		GoodsSn:         goods.GoodsSn,
		ClickNum:        goods.ClickNum,
		SoldNum:         goods.SoldNum,
		FavNum:          goods.FavNum,
		MarketPrice:     goods.MarketPrice,
		ShopPrice:       goods.ShopPrice,
		GoodsBrief:      goods.GoodsBrief,
		ShipFree:        goods.ShipFree,
		GoodsFrontImage: goods.GoodsFrontImages,
		IsNew:           goods.IsNew,
		IsHot:           goods.IsHot,
		OnSale:          goods.OnSale,
		DescImages:      goods.DescImages,
		Images:          goods.Images,
		Category: &goodspb.CategoryBriefInfoResponse{
			Id:   goods.Category.ID,
			Name: goods.Category.Name,
		},
		Brand: &goodspb.BrandInfoResponse{
			Id:   goods.Brands.ID,
			Name: goods.Brands.Name,
			Logo: goods.Brands.Logo,
		},
	}
	return goodsInfoRsp
}

func (s *GoodsService) GoodsList(c context.Context, req *goodspb.GoodsFilterRequest) (*goodspb.GoodsListResponse, error) {
	goodsListRsp := &goodspb.GoodsListResponse{}
	q := elastic.NewBoolQuery()
	localDB := global.DB.Model(model.Goods{})
	if req.KeyWords != "" {
		q.Must(elastic.NewMultiMatchQuery(req.KeyWords, "name", "goods_brief"))
	}
	if req.IsHot {
		//localDB = localDB.Where("is_hot=true")
		q.Filter(elastic.NewTermQuery("is_hot", req.IsHot))
	}
	if req.IsNew {
		//localDB = localDB.Where(model.Goods{IsNew: true})
		//must会计算分数，filter不会计算分数，新品和热销两个字段不用参加算分
		//q.Must(elastic.NewTermQuery("is_new",req.IsNew))
		q.Filter(elastic.NewTermQuery("is_new", req.IsNew))
	}
	if req.PriceMin > 0 {
		//localDB = localDB.Where("shop_price >= ?",req.PriceMin)
		q.Filter(elastic.NewRangeQuery("shop_price").Gte(req.PriceMin))
	}
	if req.PriceMax > 0 {
		//localDB = localDB.Where("shop_price <= ?",req.PriceMax)
		q.Filter(elastic.NewRangeQuery("shop_price").Lte(req.PriceMin))
	}
	if req.Brand > 0 {
		//localDB = localDB.Where("brand_id = ?",req.Brand)
		q.Filter(elastic.NewTermQuery("brand_id", req.Brand))
	}
	//通过分类category查询
	subQuery := ""
	categoryIds := make([]interface{}, 0)
	if req.TopCategory > 0 {
		var category model.Category
		if result := global.DB.First(&category, req.TopCategory); result.RowsAffected == 0 {
			return nil, status.Error(codes.NotFound, "商品分类不存在")
		}
		if category.Level == 1 {
			subQuery = fmt.Sprintf("select id from category where parent_category_id in (select id from category WHERE parent_category_id = %d", req.TopCategory)
		} else if category.Level == 2 {
			subQuery = fmt.Sprintf("select id from category WHERE parent_category_id = %d", req.TopCategory)
		} else if category.Level == 3 {
			subQuery = fmt.Sprintf("selecet id from category WHERE id = %d", req.TopCategory)
		}
		type result struct {
			id int32
		}
		var results []result
		global.DB.Model(&model.Category{}).Raw(subQuery).Scan(&results)
		for _, res := range results {
			categoryIds = append(categoryIds, res.id)
		}
		//terms查询
		q = q.Filter(elastic.NewTermsQuery("category_id",categoryIds...))
		//localDB = localDB.Where(fmt.Sprintf("category_id in (%s)",subQuery))
	}


	//分页
	if req.Pages == 0{
		req.Pages = 1
	}
	switch {
	case req.PagePerNums > 100:
		req.PagePerNums = 100
	case req.PagePerNums < 0:
		req.PagePerNums = 10
	}
	//去es进行搜索
	result,err := global.EsClient.Search().Index(model.EsGoods{}.GetIndexName()).Query(q).From(int(req.Pages)).Size(int(req.PagePerNums)).Do(context.Background())
	if err != nil{
		return nil, err
	}
	goodsListRsp.Total = int32(result.Hits.TotalHits.Value)
	goodsids := make([]int32,0)
	for _,v := range result.Hits.Hits {
		goods := model.EsGoods{}
		_ = json.Unmarshal(v.Source,&goods)
		goodsids = append(goodsids,goods.ID)
	}
	var goods []model.Goods
	if res := localDB.Preload("Category").Preload("Brands").Find(&goods,goodsids); result.Error != nil {
		return nil, res.Error
	}
	for _, good := range goods {
		goodsInfoRsp := Model2Rsp(good)
		goodsListRsp.Data = append(goodsListRsp.Data, &goodsInfoRsp)
	}
	zap.S().Info(goodsListRsp)
	return goodsListRsp, nil
}

//现在用户提交订单有多个商品，你得批量查询商品的信息吧
func (s *GoodsService) BatchGetGoods(c context.Context, req *goodspb.BatchGoodsIdInfo) (*goodspb.GoodsListResponse, error) {
	var goods []model.Goods
	goodsListRsp := &goodspb.GoodsListResponse{}
	result := global.DB.Preload("Category").Preload("Brands").Find(&goods, req.Id)
	for _, good := range goods {
		goodsInfoRsp := Model2Rsp(good)
		goodsListRsp.Data = append(goodsListRsp.Data, &goodsInfoRsp)
	}
	goodsListRsp.Total = int32(result.RowsAffected)
	return goodsListRsp, nil
}
func (s *GoodsService) CreateGoods(c context.Context, req *goodspb.CreateGoodsInfo) (*goodspb.GoodsInfoResponse, error) {
	var category model.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "商品分类不存在")
	}
	var brand model.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}
	good := model.Goods{
		Brands:           brand,
		BrandsID:         brand.ID,
		Category:         category,
		CategoryID:       category.ID,
		Name:             req.Name,
		GoodsSn:          req.GoodsSn,
		MarketPrice:      req.MarketPrice,
		ShopPrice:        req.ShopPrice,
		GoodsBrief:       req.GoodsBrief,
		ShipFree:         req.ShipFree,
		Images:           req.Images,
		DescImages:       req.DescImages,
		GoodsFrontImages: req.GoodsFrontImage,
		IsNew:            req.IsNew,
		IsHot:            req.IsHot,
		OnSale:           req.OnSale,
	}
	tx := global.DB.Begin()
	//save会调用afterCreate函数
	res := tx.Save(&good)
	if res.Error != nil{
		tx.Rollback()
		return nil,res.Error
	}
	tx.Commit()
	return &goodspb.GoodsInfoResponse{
		Id: good.ID,
	}, nil
}
func (s *GoodsService) DeleteGoods(c context.Context, req *goodspb.DeleteGoodsInfo) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Goods{BaseModel: model.BaseModel{ID: req.Id}}, req.Id); result.Error != nil {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	return &emptypb.Empty{}, nil
}

func (s *GoodsService) UpdateGoods(c context.Context, req *goodspb.CreateGoodsInfo) (*emptypb.Empty, error) {
	var goods model.Goods

	if result := global.DB.First(&goods, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	var category model.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "商品分类不存在")
	}
	var brand model.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}
	goods.Brands = brand
	goods.BrandsID = brand.ID
	goods.Category = category
	goods.CategoryID = category.ID
	goods.Name = req.Name
	goods.GoodsSn = req.GoodsSn
	goods.MarketPrice = req.MarketPrice
	goods.ShopPrice = req.ShopPrice
	goods.GoodsBrief = req.GoodsBrief
	goods.ShipFree = req.ShipFree
	goods.Images = req.Images
	goods.DescImages = req.DescImages
	goods.GoodsFrontImages = req.GoodsFrontImage
	goods.IsNew = req.IsNew
	goods.IsHot = req.IsHot
	goods.OnSale = req.OnSale
	global.DB.Save(&goods)
	return &emptypb.Empty{}, nil

}
func (s *GoodsService) GetGoodsDetail(c context.Context, req *goodspb.GoodInfoRequest) (*goodspb.GoodsInfoResponse, error) {
	var good model.Goods
	if result := global.DB.Preload("Category").Preload("Brands").First(&good, req.Id); result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "商品不存在")
	}
	goodsInfoRsp := Model2Rsp(good)
	return &goodsInfoRsp, nil
}

//
//
