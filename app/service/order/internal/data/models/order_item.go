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
	// SKU ID
	SkuID int64 `json:"sku_id" gorm:"index:idx_order_item_sku_id;not null;default:0;comment:SKU ID"`
	// 商品名称
	ProductName string `json:"product_name" gorm:"type:varchar(128);not null;comment:商品名称"`
	// 规格描述快照（下单时SKU的规格描述，如"红色/大号"）
	SpecDesc string `json:"spec_desc" gorm:"type:varchar(256);default:'';comment:规格描述快照"`
	// 商品图片（下单时快照：优先取SKU图片，无则取SPU主图）
	Image string `json:"image" gorm:"type:varchar(1024);default:'';comment:商品图片"`
	// 单价（单位：分）
	Price int64 `json:"price" gorm:"not null;default:0;comment:单价"`
	// 数量
	Quantity int32 `json:"quantity" gorm:"not null;default:1;comment:数量"`
}

type IOrderItem interface {
}
