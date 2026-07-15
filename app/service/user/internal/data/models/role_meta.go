package models

import "gorm.io/gorm"

// RoleMeta 角色元数据（名称、描述），与 Casbin 策略分离存储
type RoleMeta struct {
	gorm.Model
	Slug string `gorm:"uniqueIndex;type:varchar(100)"`
	Name string `gorm:"type:varchar(100)"`
	Desc string `gorm:"type:varchar(200)"`
}

type IRoleMeta interface {
}