package models

import (
	"cyi-app/pkg/util/gorm"
)

// OrderItem 订单商品明细表
type OrderItem struct {
	gorm.Model
	// 订单ID
	OrderID int64 `json:"order_id" gorm:"index;not null;comment:订单ID"`
	// 商品ID
	ProductID int64 `json:"product_id" gorm:"not null;comment:商品ID"`
	// 商品名称
	ProductName string `json:"product_name" gorm:"type:varchar(128);not null;comment:商品名称"`
	// 单价（单位：分）
	Price int64 `json:"price" gorm:"not null;default:0;comment:单价"`
	// 数量
	Quantity int32 `json:"quantity" gorm:"not null;default:1;comment:数量"`
}

type IOrderItem interface {
}
