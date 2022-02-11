package main

import (
	"awesomeProject/shengxian/mxshop_srv/inventory_srv/model"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func main() {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := "root:root@tcp(127.0.0.1:3306)/mxshop_inventory_srv?charset=utf8mb4&parseTime=True&loc=Local"
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
	//_ = db.AutoMigrate(&model.Inventory{},&model.StockSellDetail{})

	//查询积分消耗情况
	//var unlock_records []model.UnlockScoresCopy1
	//db.Where("create_time BETWEEN ? AND ?",1640966400,1643644799).Find(&unlock_records)
	//fmt.Println(len(unlock_records))
	//var total int
	//for record := range unlock_records{
	//	//fmt.Println(unlock_records[record].ScoresID)
	//	scoreId := unlock_records[record].ScoresID
	//	var score model.Scores
	//	db.Where(model.Scores{
	//		ID: scoreId,
	//	}).First(&score)
	//	if score.XmlPath != ""{
	//		fmt.Println(score.XmlPath)
	//		total+=300
	//	}else if score.AudioPath != ""{
	//		total += 200
	//	} else {
	//		total += 50
	//	}
	//}
	//fmt.Println(total)


	//查询积分充值金额
	//var vipRecord []model.VipRecord
	//db.Model(model.VipRecord{}).Where("content LIKE ? AND create_time BETWEEN ? AND ?","%购买%",1640966400,1643644799).Find(&vipRecord)
	//fmt.Println(len(vipRecord))
	//num := 0
	//amount := 0
	//for i := range vipRecord{
	//	if vipRecord[i].Content == "购买800积分"{
	//		num += 800
	//		amount += 8
	//	}else if vipRecord[i].Content == "购买1800积分"{
	//		num += 1800
	//		amount += 18
	//	}else if vipRecord[i].Content == "购买9800积分"{
	//		amount += 98
	//		num += 9800
	//	}
	//
	//}
	//fmt.Println(amount)
	//fmt.Println(num)

	//查询积分消耗个数
	//var vipRecords []model.VipRecord
	//db.Model(model.VipRecord{}).Where("content LIKE ? AND create_time BETWEEN ? AND ?","%解锁%积分%",1640966400,1643644799).Find(&vipRecords)
	//fmt.Println(len(vipRecords))
	//total := 0
	//for i := range vipRecords{
	//	stringSlice := strings.Split(vipRecords[i].Content,":")
	//	//fmt.Println(stringSlice[len(stringSlice)-1])
	//	starString := stringSlice[len(stringSlice)-1]
	//	if starString == "50积分"{
	//		total += 50
	//	}else if starString == "300积分"{
	//		total += 300
	//	}else if starString == "200积分"{
	//		total += 200
	//	}else {
	//		fmt.Println(starString)
	//	}
	//
	//}
	//fmt.Println(total)

	//查询积分赠送个数
	//var vipRecords []model.VipRecord
	//
	//total := 0
	//res0 := db.Model(model.VipRecord{}).Where("content LIKE ? AND create_time BETWEEN ? AND ?","%20积分%",1640966400,1643644799).Find(&vipRecords)
	//res1 := db.Model(model.VipRecord{}).Where("content LIKE ? AND create_time BETWEEN ? AND ?","%100积分%",1640966400,1643644799).Find(&vipRecords)
	//res2 := db.Model(model.VipRecord{}).Where("content LIKE ? AND create_time BETWEEN ? AND ?","%500积分%",1640966400,1643644799).Find(&vipRecords)
	//total += int(res0.RowsAffected * 20) + int(res1.RowsAffected * 100) + int(res2.RowsAffected * 500)
	//
	//fmt.Println(total)


	//查询微信充值金额
}
