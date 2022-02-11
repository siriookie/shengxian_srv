package initialize

import (
	"awesomeProject/shengxian/mxshop_srv/inventory_srv/global"
	"fmt"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

func InitRedSync(){
	client := goredislib.NewClient(&goredislib.Options{
		Addr: fmt.Sprintf("%s:%d",global.ServerConfig.RedisInfo.Host, &global.ServerConfig.RedisInfo),
	})
	global.RedisPool = goredis.NewPool(client) // or, pool := redigo.NewPool(...)

}
