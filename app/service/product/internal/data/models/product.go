package models

import (
	"cyi-app/pkg/util/gorm"
)

type Product struct {
	gorm.Model
	// 商品名称
	Name string `json:"name" gorm:"type:varchar(128);not null;comment:商品名称"`
	// 商品描述
	Description string `json:"description" gorm:"type:varchar(512);default:'';comment:商品描述"`
	// 价格（单位：分）
	Price int64 `json:"price" gorm:"not null;default:0;comment:价格，单位分"`
	// 状态 0-下架 1-上架
	Status int32 `json:"status" gorm:"not null;default:1;comment:状态 0-下架 1-上架"`
}

type IProduct interface {
}
