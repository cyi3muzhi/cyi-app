package models

import (
	"cyi-app/pkg/util/gorm"
)

type Stock struct {
	gorm.Model
	// 商品ID
	ProductID int64 `json:"product_id" gorm:"index:idx_product_id;not null;default:0;comment:商品ID"`
	// SKU ID（为0时表示 product 级库存）
	SkuID int64 `json:"sku_id" gorm:"index:idx_sku_id;not null;default:0;comment:SKU ID"`
	// 总库存
	Total int32 `json:"total" gorm:"not null;default:0;comment:总库存数量"`
	// 可用库存
	Available int32 `json:"available" gorm:"not null;default:0;comment:可用库存数量"`
	// 预留库存（待支付订单锁定）
	Reserved int32 `json:"reserved" gorm:"not null;default:0;comment:预留库存数量"`
}

type IStock interface {
}
