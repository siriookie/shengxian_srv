package main

import (
	"awesomeProject/shengxian/mxshop_srv/goods_srv/model"
	"context"
	"github.com/olivere/elastic/v7"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {

	//_ = db.AutoMigrate(&model.Category{},&model.Banner{},&model.Brands{},&model.Goods{},&model.GoodsCategoryBrand{})
	///*
	//生成十个用户
	// */

}

func Mysql2Es(){
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := "root:root@tcp(127.0.0.1:3306)/mxshop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        // 禁用彩色打印
		},
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}
	logger1 := log.New(os.Stdout,"es-test",log.LstdFlags)
	client, err := elastic.NewClient(elastic.SetSniff(false),elastic.SetTraceLog(logger1))
	if err != nil {
		// Handle error
		panic(err)
	}
	var goods []model.Goods
	db.Find(&goods)
	for _,g := range goods{
		esModel := model.EsGoods{
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
		_,err = client.Index().Index(esModel.GetIndexName()).BodyJson(esModel).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
		if err != nil{
			panic(err)
		}
	}
}