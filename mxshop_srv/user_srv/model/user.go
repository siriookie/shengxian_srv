package model

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID         int32     `gorm:"primarykey"`
	CreateAt   time.Time `gorm:"create_at"`
	UpdateTime time.Time `gorm:"update_time"`
	DeleteAt   gorm.DeletedAt
	IsDeleted  bool
}

type User struct {
	BaseModel
	Mobile   string     `gorm:"index:idx_mobile;unique;type:varchar(11);not null"`
	Password string     `gorm:"type:varchar(512);not null"`
	NickName string     `gorm:"type:varchar(20)"`
	Birthday *time.Time `gorm:"type:datetime"`
	Gender   string     `gorm:"column:gender;default:male;type:varchar(6) comment 'female表示女，male表示男'"`
	Role     int        `gorm:"column:role;default:1;type:int"`
}
