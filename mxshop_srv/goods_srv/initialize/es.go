package initialize

import (
	"awesomeProject/shengxian/mxshop_srv/goods_srv/global"
	"awesomeProject/shengxian/mxshop_srv/goods_srv/model"
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
)

func InitEs() {
	//初始化连接
	host := fmt.Sprintf("http://%s:%d", global.ServerConfig.ESInfo.Host, global.ServerConfig.ESInfo.Port)
	logger := log.New(os.Stdout,"es-test",log.LstdFlags)
	var err error
	global.EsClient, err = elastic.NewClient(elastic.SetURL(host),elastic.SetSniff(false),elastic.SetTraceLog(logger))
	if err != nil {
		// Handle error
		panic(err)
	}
	//新建mapping
	exist, err := global.EsClient.IndexExists(model.EsGoods{}.GetIndexName()).Do(context.Background())
	if err != nil{
		panic(err)
	}
	if !exist{
		_,err = global.EsClient.CreateIndex(model.EsGoods{}.GetIndexName()).BodyString(model.EsGoods{}.GetMapping()).Do(context.Background())
		if err != nil{
			panic(err)
		}
	}
}
