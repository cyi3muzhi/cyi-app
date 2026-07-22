package models

import (
	"cyi-app/pkg/util/gorm"
)

type Product struct {
	gorm.Model
	// 商品名称
	Name string `json:"name" gorm:"type:varchar(128);not null;comment:商品名称"`
	// 商品描述
	Description string `json:"description" gorm:"type:text;comment:商品描述"`
	// 价格（单位：分）
	Price int64 `json:"price" gorm:"not null;default:0;comment:价格，单位分"`
	// 状态 0-下架 1-上架
	Status int32 `json:"status" gorm:"not null;default:0;comment:状态 0-下架 1-上架"`
	// 商品分类ID
	CategoryID int64 `json:"category_id" gorm:"index;not null;default:0;comment:商品分类ID"`
	// 商品图片
	Image string `json:"image" gorm:"type:varchar(1024);default:'';comment:商品图片（支持完整 URL 和 OSS 路径）"`
	// 轮播图JSON数组
	Images string `json:"images" gorm:"type:text;comment:轮播图JSON数组" mapstructure:"-"`
}

type IProduct interface {
}
