package models

import (
	"cyi-app/pkg/util/gorm"
)

type CouponDistRecord struct {
	gorm.Model
	CouponID  int64  `json:"coupon_id" gorm:"index;not null;comment:关联优惠券ID"`
	UserID    int64  `json:"user_id" gorm:"index;not null;comment:领取用户ID"`
	Source    string `json:"source" gorm:"comment:领取来源"`
	Status    int32  `json:"status" gorm:"default:0;comment:状态:0-未使用 1-已使用 2-已过期"`
	UsedTime  int64  `json:"used_time" gorm:"default:0;comment:使用时间"`
	OrderID   int64  `json:"order_id" gorm:"default:0;comment:核销订单ID"`
}

type ICouponDistRecord interface {
}