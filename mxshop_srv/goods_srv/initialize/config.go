package initialize

import (
	"awesomeProject/shengxian/mxshop_srv/goods_srv/global"
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func GetEnvInfo(env string) bool{
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

func InitConfig(){
	debug := GetEnvInfo("MXSHOP_DEBUG")
	zap.S().Info("is debug env？",debug,"\n")
	configFilePrefix := "config"
	configFileName := fmt.Sprintf("goods_srv/%s-pro.yaml",configFilePrefix)
	if debug{
		configFileName = fmt.Sprintf("goods_srv/%s-debug.yaml",configFilePrefix)
	}
	v := viper.New()
	//文件路径如何设置
	v.SetConfigFile(configFileName)
	if err := v.ReadInConfig();err != nil{
		panic(err)
	}
	if err := v.Unmarshal(global.NacosConfig);err != nil{
		panic(err)
	}
	zap.S().Debug(global.NacosConfig)

	//从nacos中拿配置信息
	sc := []constant.ServerConfig{
		{
			IpAddr:      global.NacosConfig.Host,
			Port:        uint64(global.NacosConfig.Port),
		},
	}

	cc := constant.ClientConfig{
		NamespaceId:         global.NacosConfig.Namespace, // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId。当namespace是public时，此处填空字符串。
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
	}

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil{
		panic(err)
	}
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: global.NacosConfig.DataId,
		Group:  global.NacosConfig.Group})
	if err != nil{
		panic(err)
	}
	//zap.S().Info(content)
	err = json.Unmarshal([]byte(content),global.ServerConfig)
	if err != nil{
		zap.S().Fatalf("读取nacos失败",err.Error())
	}
	zap.S().Info("serverInfo为",global.ServerConfig)
}
