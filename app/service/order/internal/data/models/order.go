package models

import (
	"cyi-app/pkg/util/gorm"
)

type Order struct {
	gorm.Model
	// 用户ID
	UserID int64 `json:"user_id" gorm:"index;not null;comment:用户ID"`
	// 订单总金额（单位：分）
	TotalPrice int64 `json:"total_price" gorm:"not null;default:0;comment:订单总金额"`
	// 状态
	Status int32 `json:"status" gorm:"not null;default:0;comment:状态"`
	// 订单号
	OrderNo string `json:"order_no" gorm:"not null;default:'';comment:订单号"`
	// 支付金额（单位：分）
	PayAmount int64 `json:"pay_amount" gorm:"not null;default:0;comment:支付金额"`
}

type IOrder interface {
}
