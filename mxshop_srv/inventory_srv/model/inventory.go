package model

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID         int32     `gorm:"primarykey;type:int"`
	CreateAt   time.Time `gorm:"create_at"`
	UpdateTime time.Time `gorm:"update_time"`
	DeleteAt   gorm.DeletedAt
	IsDeleted  bool
}

type GoodsDetail struct {
	Goods int32
	Num   int32
}

type GoodsDetailList []GoodsDetail //这样就会存一个字符串类型的json到数据库里

func (g *GoodsDetailList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), g)
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (g GoodsDetailList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

type Inventory struct {
	BaseModel
	Goods   int32 `gorm:"type:int;index"`
	Stocks  int32 `gorm:"type:int"`
	Version int32 `gorm:"type:int"` //涉及到后面的分布式锁
}

//type DeliveryHistory struct {
//	Goods int32 `gorm:"type:int;index"`
//	Nums int32	`gorm:"type:int"`
//	OrderSn string `gorm:"type:varchar(200)"`
//	Status string`gorm:"type:varchar(200)"`	//如果是1代表已扣减，如果是2代表已归还
//}

type StockSellDetail struct {
	OrderSn string          `gorm:"type:varchar(200);index:idx_ordersn,unique;"`
	Status  int32           `gorm:"type:int"`          //如果是1代表已扣减，如果是2代表已归还
	Detail  GoodsDetailList `gorm:"type:varchar(200)"` //记录订单中扣减商品或者归还商品的数量
}

type Scores struct {
	ID        int32  `gorm:"primarykey;type:int"`
	XmlPath   string `gorm:"xml_path;type:varchar(100)"`
	AudioPath string `gorm:"audio_path;type:varchar(100)"`
}

type UnlockScoresCopy1 struct {
	ID        int32  `gorm:"primarykey;type:int"`
	CreateTime   int32 `gorm:"create_time;type:int"`
	ScoresID int32 `gorm:"scores_id;type:int"`
}

type VipRecord struct{
	ID        int32  `gorm:"primarykey;type:int"`
	CreateTime   int32 `gorm:"create_time;type:int"`
	Content string`gorm:"type:varchar(100)"`
}