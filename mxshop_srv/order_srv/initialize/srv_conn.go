package initialize

import (
	"awesomeProject/shengxian/mxshop_srv/order_srv/global"
	goodspb "awesomeProject/shengxian/mxshop_srv/order_srv/proto/goods"
	invpb "awesomeProject/shengxian/mxshop_srv/order_srv/proto/inventory"
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)
//	InitSrvConn 用来对grpc客户端进行初始化，做了负载均衡，轮询方式
func InitSrvConn(){
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s",
			global.ServerConfig.ConsulInfo.Host,
			global.ServerConfig.ConsulInfo.Port,
			global.ServerConfig.GoodsSrvInfo.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("[InitSrvConn] 商品服务连接失败","msg",err.Error())
	}
	GoodsSrvClient := goodspb.NewGoodsClient(goodsConn)
	global.GoodsSrvClient = GoodsSrvClient


	invConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s",
			global.ServerConfig.ConsulInfo.Host,
			global.ServerConfig.ConsulInfo.Port,
			global.ServerConfig.InventorySrvInfo.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("[InitSrvConn] 库存服务连接失败","msg",err.Error())
	}
	InvSrvClient := invpb.NewInventoryClient(invConn)
	global.InventorySrvClient = InvSrvClient
}

