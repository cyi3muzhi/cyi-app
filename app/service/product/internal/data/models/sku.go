package models

import "cyi-app/pkg/util/gorm"

// Sku SKU GORM 模型
// 每个SKU对应一个具体的规格组合，如"颜色:红色,尺码:XL"
type Sku struct {
	gorm.Model
	// 商品ID
	ProductID int64 `json:"product_id" gorm:"index;not null;comment:商品ID"`
	// SKU编码
	SkuCode string `json:"sku_code" gorm:"type:varchar(64);default:'';comment:SKU编码"`
	// 价格（单位：分）
	Price int64 `json:"price" gorm:"not null;default:0;comment:价格(分)"`
	// 规格描述 JSON，如 {"颜色":"红色","尺码":"XL"}
	SpecDesc string `json:"spec_desc" gorm:"type:varchar(256);default:'';comment:规格描述JSON"`
	// SKU图片
	Image string `json:"image" gorm:"type:varchar(1024);default:'';comment:SKU图片"`
	// 销量
	Sales int32 `json:"sales" gorm:"not null;default:0;comment:SKU级别销量"`
}

// ISku GORM gen 接口标记
type ISku interface{}
