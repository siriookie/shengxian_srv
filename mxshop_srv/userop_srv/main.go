package main

import (
	"awesomeProject/shengxian/mxshop_srv/userop_srv/global"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/handler"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/initialize"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/proto/addr"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/proto/message"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/proto/userfav"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/utils"
	"flag"
	"fmt"
	"github.com/google/uuid"

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

	//初始化redsync
	initialize.InitRedSync()
	if *port == 0{
		var err error
		*port, err = utils.GetFreePort()
		if err != nil {
			panic(err.Error())
		}
	}

	server := grpc.NewServer()
	userfav.RegisterUserFavServer(server,&handler.UserOpServer{})
	addr.RegisterAddressServer(server,&handler.UserOpServer{})
	message.RegisterMessageServer(server,&handler.UserOpServer{})
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
		GRPC: fmt.Sprintf("%s:%d","192.168.0.102",*port),
		Timeout: "5s",
		Interval: "5s",
		DeregisterCriticalServiceAfter: "10s",
	}
	//生成注册对象
	registration := new(api.AgentServiceRegistration)
	registration.Name = global.ServerConfig.Name
	registration.ID = serviceID.String()
	registration.Port = *port
	registration.Tags = []string{"mxshop","nmsl","userop"}
	registration.Address = "192.168.0.102"
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



	quit := make(chan os.Signal)
	signal.Notify(quit,syscall.SIGINT,syscall.SIGTERM)
	<- quit
	if err = client.Agent().ServiceDeregister(serviceID.String());err != nil{
		zap.S().Error("注销失败")
	}
	zap.S().Info("注销成功")

}
