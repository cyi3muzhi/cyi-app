package models

import "cyi-app/pkg/util/gorm"

// SpecGroup 规格组 GORM 模型
// 对应商品的一组规格名称，如"颜色"、"尺码"
type SpecGroup struct {
	gorm.Model
	// 商品ID
	ProductID int64 `json:"product_id" gorm:"index;not null;comment:商品ID"`
	// 规格名称（如颜色、尺码）
	Name string `json:"name" gorm:"type:varchar(64);not null;comment:规格名称(如颜色)"`
}

// ISpecGroup GORM gen 接口标记
type ISpecGroup interface{}
