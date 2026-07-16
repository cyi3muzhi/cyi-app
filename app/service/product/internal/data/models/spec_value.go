package models

import "cyi-app/pkg/util/gorm"

// SpecValue 规格值 GORM 模型
// 对应规格组下的具体取值，如"红色"、"XL"
type SpecValue struct {
	gorm.Model
	// 规格组ID
	SpecGroupID int64 `json:"spec_group_id" gorm:"index;not null;comment:规格组ID"`
	// 规格值（如红色、XL）
	Value string `json:"value" gorm:"type:varchar(128);not null;comment:规格值(如红色)"`
}

// ISpecValue GORM gen 接口标记
type ISpecValue interface{}
