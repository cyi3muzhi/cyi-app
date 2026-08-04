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
	// 支付金额（单位：分）
	PayAmount int64 `json:"pay_amount" gorm:"not null;default:0;comment:支付金额"`
	// 运费（单位：分）
	Freight int64 `json:"freight" gorm:"not null;default:0;comment:运费（单位：分）"`
	// 优惠金额（单位：分）
	DiscountAmount int64 `json:"discount_amount" gorm:"not null;default:0;comment:优惠金额（单位：分）"`
	// 状态（与 api/service/order/v1 OrderStatus 枚举一致）: 1-待付款 2-已支付 3-待发货 4-已发货 5-待收货 6-已完成 7-已取消 8-退款中 9-已退款
	Status int32 `json:"status" gorm:"not null;default:0;comment:状态: 1-待付款 2-已支付 3-待发货 4-已发货 5-待收货 6-已完成 7-已取消 8-退款中 9-已退款"`
	// 订单号
	OrderNo string `json:"order_no" gorm:"uniqueIndex;not null;default:'';comment:订单号"`
	// 优惠券发放记录ID
	CouponRecordID int64 `json:"coupon_record_id" gorm:"default:0;comment:优惠券发放记录ID"`
}

type IOrder interface {
}
