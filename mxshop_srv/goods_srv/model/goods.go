package model

import (
	"awesomeProject/shengxian/mxshop_srv/goods_srv/global"
	"context"
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type BaseModel struct {
	ID         int32     `gorm:"primarykey;type:int"`
	CreateAt   time.Time `gorm:"create_at"`
	UpdateTime time.Time `gorm:"update_time"`
	DeleteAt   gorm.DeletedAt
	IsDeleted  bool
}



type GormList []string

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte),g)
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

type Category struct {
	BaseModel
	Name             string `gorm:"type:varchar(20);not null"`
	Level            int32  `gorm:"type:int;not null;default:1"` //	默认是一级类目
	IsTab            bool   `gorm:"default:false;not null"`
	ParentCategoryID int32
	ParentCategory   *Category
	SubCategory []*Category `gorm:"foreignKey:ParentCategoryID;references:ID"`
}

type Brands struct {
	BaseModel
	Name string `gorm:"type:varchar(20);not null"`
	Logo string `gorm:"type:varchar(200);not null;default:''"`
}

//索引名一样就会建立联合索引
type GoodsCategoryBrand struct {
	BaseModel
	CategoryID int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Category   Category

	Brands   Brands
	BrandsID int32 `gorm:"type:int;index:idx_category_brand,unique"`
}

type Banner struct {
	BaseModel
	Image string `gorm:"type:varchar(200);not null"`
	Url   string `gorm:"type:varchar(200);not null"`
	Index int32 `gorm:"type:int;not null;default:1"`
}

type Goods struct {
	BaseModel
	Category Category
	CategoryID int32 `gorm:"type:int;not null"`

	Brands   Brands
	BrandsID int32 `gorm:"type:int;not null"`

	OnSale bool `gorm:"not null;default:false"`
	ShipFree bool `gorm:"not null;default:false"`
	IsNew bool `gorm:"not null;default:false"`
	IsHot bool `gorm:"not null;default:false"`

	Name string `gorm:"type:varchar(50);not null"`
	GoodsSn string `gorm:"type:varchar(50);not null"`
	ClickNum int32 `gorm:"type:int;default:0;not null"`
	SoldNum int32 `gorm:"type:int;default:0;not null"`
	FavNum int32 `gorm:"type:int;default:0;not null"`
	MarketPrice float32 `gorm:"not null"`
	ShopPrice float32 `gorm:"not null"`
	GoodsBrief string `gorm:"type:varchar(50);not null"`	//商品简介
	Images GormList `gorm:"type:varchar(1000);not null"`	//商品简介页图片
	DescImages GormList `gorm:"type:varchar(1000);not null"`	//商品详情页图片
	GoodsFrontImages string `gorm:"type:varchar(200);not null"`	//封面图片
}

func (g *Goods)AfterCreate(tx *gorm.DB)(err error)  {
	esModel := EsGoods{
		ID: g.ID,
		CategoryID: g.CategoryID,
		BrandsID: g.BrandsID,
		OnSale: g.OnSale,
		ShipFree: g.ShipFree,
		IsHot: g.IsHot,
		IsNew: g.IsNew,
		Name: g.Name,
		ClickNum: g.ClickNum,
		SoldNum: g.SoldNum,
		FavNum: g.FavNum,
		MarketPrice: g.MarketPrice,
		ShopPrice: g.ShopPrice,
		GoodsBrief: g.GoodsBrief,
	}
	_,err = global.EsClient.Index().Index(esModel.GetIndexName()).BodyJson(esModel).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
	if err != nil{
		return err
	}
	return nil
}

