package main

import (
	"awesomeProject/shengxian/mxshop_srv/user_srv/model"
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func main() {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := "root:root@tcp(127.0.0.1:3306)/mxshop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"
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
	// _ = db.AutoMigrate(&model.User{})
	/*
	生成十个用户
	 */
	options := &password.Options{16,100,32,sha512.New}
	salt,encodePwd := password.Encode("123",options)
	newPwd := fmt.Sprintf("$pbkdf2-sha512$%s$%s",salt,encodePwd)
	for i:=0;i<10;i++{
		user := model.User{
			NickName: fmt.Sprintf("test%d",i),
			Mobile: fmt.Sprintf("1528600566%d",i),
			Password: newPwd,

		}
		db.Save(&user)
	}
}
