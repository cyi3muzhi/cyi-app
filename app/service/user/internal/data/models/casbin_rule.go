package models

import "gorm.io/gorm"

// CasbinRule Casbin 策略规则表
type CasbinRule struct {
	gorm.Model
	Ptype string `gorm:"column:ptype;type:varchar(100);not null;index"`
	V0    string `gorm:"column:v0;type:varchar(100);default:''"`
	V1    string `gorm:"column:v1;type:varchar(100);default:''"`
	V2    string `gorm:"column:v2;type:varchar(100);default:''"`
	V3    string `gorm:"column:v3;type:varchar(100);default:''"`
	V4    string `gorm:"column:v4;type:varchar(100);default:''"`
	V5    string `gorm:"column:v5;type:varchar(100);default:''"`
}

type ICasbinRule interface {
}
