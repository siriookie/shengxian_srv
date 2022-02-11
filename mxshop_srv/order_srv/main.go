package main

import (
	"awesomeProject/shengxian/mxshop_srv/order_srv/global"
	"awesomeProject/shengxian/mxshop_srv/order_srv/handler"
	"awesomeProject/shengxian/mxshop_srv/order_srv/initialize"
	orderpb "awesomeProject/shengxian/mxshop_srv/order_srv/proto/order"
	"awesomeProject/shengxian/mxshop_srv/order_srv/utils"
	"flag"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/google/uuid"
	"time"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ip := flag.String("ip","0.0.0.0","ip地址")
	port := flag.Int("port",0,"端口号")

	flag.Parse()
	//初始化日志
	initialize.InitLogger()
	//初始化配置
	initialize.InitConfig()
	//初始化DB
	initialize.InitDB()
	//初始化别的微服务的连接
	initialize.InitSrvConn()
	//获取当前ipv4地址,用来在consul中注册服务
	initialize.InitLocalIPv4Address()
	if *port == 0{
		var err error
		*port, err = utils.GetFreePort()
		if err != nil {
			panic(err.Error())
		}
	}

	server := grpc.NewServer()
	orderpb.RegisterOrderServer(server,&handler.OrderService{})
	lis, err := net.Listen("tcp",fmt.Sprintf("%s:%d",*ip,*port))
	if err != nil{
		panic("failed to listen port"+err.Error())
	}
	//注册健康检查服务(grpc自己提供的，本项目中是去配置consul来检查健康)
	grpc_health_v1.RegisterHealthServer(server,health.NewServer())
	serviceID := uuid.New()
	//grpc服务注册到consul
	cfg := api.DefaultConfig()
	cfg.Address  = fmt.Sprintf("%s:%d",global.ServerConfig.ConsulInfo.Host,global.ServerConfig.ConsulInfo.Port)
	client, err := api.NewClient(cfg)
	if err != nil{
		panic(err)
	}
	//健康检查的配置
	check := &api.AgentServiceCheck{
		GRPC: fmt.Sprintf("%s:%d",global.Ipv4Addr,*port),
		Timeout: "5s",
		Interval: "5s",
		DeregisterCriticalServiceAfter: "10s",
	}
	//生成注册对象
	registration := new(api.AgentServiceRegistration)
	registration.Name = global.ServerConfig.Name
	registration.ID = serviceID.String()
	registration.Port = *port
	registration.Tags = []string{"mxshop","nmsl","order"}
	registration.Address = global.Ipv4Addr
	registration.Check = check
	if err = client.Agent().ServiceRegister(registration); err != nil{
		panic(err)
	}

	go func() {
		err = server.Serve(lis)
		if err != nil{
			panic("failed to start grpc"+err.Error())
		}
	}()
	//监听订单超时topic
	//两种consumer Pull是客户端不停的向服务器拉取数据，轮询会耗费很多服务器资源 Push是服务器有数据之后会推过来
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"127.0.0.1:9876"}), //等价下面的代码,下面的方法有点老
		//producer.WithNsResolver(primitive.NewPassthroughResolver([]string{"127.0.0.1:9876"})),
		consumer.WithGroupName("mxshop-order"), //rockctmq会对同一组的消费者进行消费偏移量的记录，所以同一组的多个消费者在消费时候不会产生重复消费，负载均衡了
	)
	if err != nil {
		panic(err)
	}
	err = c.Subscribe("order_timeout",consumer.MessageSelector{},handler.OrderTimeout)
	if err != nil{
		fmt.Println("读取消息失败")
	}
	_ = c.Start()
	defer c.Shutdown()
	time.Sleep(time.Second * 300)

	quit := make(chan os.Signal)
	signal.Notify(quit,syscall.SIGINT,syscall.SIGTERM)
	<- quit
	if err = client.Agent().ServiceDeregister(serviceID.String());err != nil{
		zap.S().Error("注销失败")
	}
	zap.S().Info("注销成功")

}
