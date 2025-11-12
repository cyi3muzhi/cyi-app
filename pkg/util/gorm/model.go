package gorm

import (
	"gorm.io/plugin/soft_delete"
)

type Model struct {
	ID         int64                 `gorm:"primarykey,autoIncrement"`
	CreateTime int64                 `gorm:"autoCreateTime"`  // 使用时间戳秒数填充创建时间
	UpdateTime int64                 `gorm:"autoUpdateTime"`  // 使用时间戳秒数填充更新时间
	IsDel      soft_delete.DeletedAt `gorm:"softDelete:flag"` // 软删除字段 0 表示未删除 1 表示已删除
}
