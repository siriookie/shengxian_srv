package global

import (
	"awesomeProject/shengxian/mxshop_srv/goods_srv/config"
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
)

//定义一个全局变量
var (
	DB *gorm.DB
	ServerConfig *config.ServerConfig =  &config.ServerConfig{}
	NacosConfig *config.NacosConfig = &config.NacosConfig{}
	EsClient *elastic.Client
)

////	init 方法在被import的时候会自动执行
//func init() {
//	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
//	dsn := "root:root@tcp(127.0.0.1:3306)/mxshop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
//	newLogger := logger.New(
//		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
//		logger.Config{
//			SlowThreshold: time.Second,   // 慢 SQL 阈值
//			LogLevel:      logger.Silent, // 日志级别
//			IgnoreRecordNotFoundError: true,   // 忽略ErrRecordNotFound（记录未找到）错误
//			Colorful:      false,         // 禁用彩色打印
//		},
//	)
//	var err error
//	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
//		Logger: newLogger,
//	})
//	if err != nil{
//		panic(err)
//	}
//}
