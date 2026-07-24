package models

import (
	"cyi-app/pkg/util/gorm"
)

type Coupon struct {
	gorm.Model
	Name        string `json:"name" gorm:"index;not null;comment:优惠券名称"`
	Type        int32  `json:"type" gorm:"not null;default:1;comment:类型:1-立减 2-折扣 3-满减"`
	Value       int64  `json:"value" gorm:"not null;default:0;comment:面值:cash/full-reduce=分;discount=百分比1-99"`
	MinAmount   int64  `json:"min_amount" gorm:"not null;default:0;comment:使用门槛（分），0=无门槛"`
	Quantity    int32  `json:"quantity" gorm:"not null;default:0;comment:计划发放上限"`
	IssuedCount int32  `json:"issued_count" gorm:"not null;default:0;comment:实际已发放数"`
	UsedCount   int32  `json:"used_count" gorm:"not null;default:0;comment:已使用数"`
	ValidStart  int64  `json:"valid_start" gorm:"not null;comment:有效期开始（Unix秒）"`
	ValidEnd    int64  `json:"valid_end" gorm:"index;not null;comment:有效期结束（Unix秒）"`
	Status      int32  `json:"status" gorm:"index;not null;default:1;comment:状态:1-active 2-pending 3-ended 4-voided"`
	Enabled     bool   `json:"enabled" gorm:"not null;default:true;comment:启用开关"`
}

type ICoupon interface {
}