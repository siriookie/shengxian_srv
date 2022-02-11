package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type BaseModel struct {
	ID         int32     `gorm:"primarykey;type:int"`
	CreateAt   time.Time `gorm:"create_at"`
	UpdateTime time.Time `gorm:"update_time"`
	IsDeleted  bool
}

type GormList []string

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), g)
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

//	ShoppingCart 购物车表
type ShoppingCart struct {
	BaseModel
	User    int32 `gorm:"type:int;index"`
	Goods   int32 `gorm:"type:int;index"` //如果没有需要根据某件商品查询订单的需求，尽量不要加这个索引，加索引1.会影响插入性能，2.会占用磁盘空间
	Nums    int32 `gorm:"type:int"`
	Checked *bool `gorm:"type:int"` //是否勾选
}

type OrderInfo struct {
	BaseModel

	User    int32  `gorm:"type:int;index"`
	OrderSn string `gorm:"type:varchar(30);index"`
	PayType string `gorm:"type:varchar(20);comment:'alipay','wechat'"`

	Status     string     `gorm:"type:varchar(30);comment:'PAYING(待支付), TRADE_SUCCESS(成功), TRADE_CLOSED(超时关闭), WAIT_BUYER_PAY(交易创建), TRADE_FINISHED(交易结束)'"`
	TradeNo    string     `gorm:"type:varchar(30);index comment:'交易号'"`
	OrderMount float32    `gorm:"comment:'交易总金额'"`
	PayTime    *time.Time `gorm:"comment:'用户支付的时间'"`

	Address      string `gorm:"type:varchar(100)"`
	SignerName   string `gorm:"type:varchar(30);comment:'收件人姓名'"`
	SignerMobile string `gorm:"type:varchar(11);comment:'收件人电话号码'"`
	Post         string `gorm:"type:varchar(50);comment:'收件人备注信息'"`
}

type OrderGoods struct {
	BaseModel

	Order int32 `gorm:"type:int;index"`
	Goods int32 `gorm:"type:int;index"`

	//有商品表的，但是还是保存了商品信息，存在明显的字段冗余，
	//1.在高并发的场景下，如果我不保存这些商品的信息，那么在我查看某个订单下的商品
	//时，就会跨服务去调用商品服务来查询商品信息，开销很大.
	//2.相当于一个订单的快照
	GoodsName  string `gorm:"type:varchar(100);index"`
	GoodsImage string `gorm:"type:varchar(200)"`
	GoodsPrice float32
	Nums       int32 `gorm:"type:int"`
}
