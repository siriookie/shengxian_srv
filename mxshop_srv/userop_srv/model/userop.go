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

type GormList []string

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), g)
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

type LeavingMessages struct{
	BaseModel

	User int32 `gorm:"type:int;index"`
	MessageType int32 `gorm:"type:int comment '留言类型: 1(留言),2(投诉),3(询问),4(售后),5(求购)'"`
	Subject string `gorm:"type:varchar(100)"`

	Message string	//不加类型默认就是text
	File string `gorm:"type:varchar(200)"`
}

type Address struct{
	BaseModel

	User int32 `gorm:"type:int;index"`
	Province string `gorm:"type:varchar(10)"`
	City string `gorm:"type:varchar(10)"`
	District string `gorm:"type:varchar(20)"`
	Address string `gorm:"type:varchar(100)"`
	SignerName string `gorm:"type:varchar(20)"`
	SignerMobile string `gorm:"type:varchar(11)"`
}

type UserFav struct{
	BaseModel
	//这两个字段要有一个联合唯一索引
	User int32 `gorm:"type:int;index:idx_user_goods,unique"`
	Goods int32 `gorm:"type:int;index:idx_user_goods,unique"`
}