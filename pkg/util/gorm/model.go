package gorm

import (
	"gorm.io/gorm"
)

// Model 基础模型，所有表嵌入此结构体
// 统一使用 deleted_at datetime(3) 软删除（NULL=未删除）
type Model struct {
	ID         int64          `gorm:"primarykey,autoIncrement"`
	CreateTime int64          `gorm:"autoCreateTime"` // 创建时间（Unix 时间戳秒）
	UpdateTime int64          `gorm:"autoUpdateTime"` // 更新时间（Unix 时间戳秒）
	DeletedAt  gorm.DeletedAt `gorm:"index"`         // 软删除时间（NULL=未删除）
}
