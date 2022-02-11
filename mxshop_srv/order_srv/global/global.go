package global

import (
	goodspb "awesomeProject/shengxian/mxshop_srv/order_srv/proto/goods"
	"awesomeProject/shengxian/mxshop_srv/order_srv/config"
	invpb "awesomeProject/shengxian/mxshop_srv/order_srv/proto/inventory"
	"gorm.io/gorm"
)

//定义一个全局变量
var (
	DB *gorm.DB
	ServerConfig *config.ServerConfig =  &config.ServerConfig{}
	NacosConfig *config.NacosConfig = &config.NacosConfig{}
	Ipv4Addr string = ""
	GoodsSrvClient goodspb.GoodsClient
	InventorySrvClient invpb.InventoryClient
)

