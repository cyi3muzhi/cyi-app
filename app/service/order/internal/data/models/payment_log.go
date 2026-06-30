package models

import (
	"cyi-app/pkg/util/gorm"
)

// PaymentLog 支付记录表
// 每次支付操作记录一行，微信/支付宝回调据此做幂等处理
type PaymentLog struct {
	gorm.Model
	// 订单ID
	OrderID int64 `json:"order_id" gorm:"index;not null;comment:订单ID"`
	// 支付方式 wechat / alipay
	PayType string `json:"pay_type" gorm:"type:varchar(20);not null;comment:支付方式"`
	// 第三方支付订单号（微信/支付宝流水号）
	TransactionID string `json:"transaction_id" gorm:"type:varchar(64);default:'';comment:第三方支付订单号"`
	// 支付金额（单位：分）
	Amount int64 `json:"amount" gorm:"not null;default:0;comment:支付金额"`
	// 支付状态 0-未支付 1-支付成功 2-支付失败 3-已退款
	PayStatus int32 `json:"pay_status" gorm:"not null;default:0;comment:支付状态"`
	// 支付结果描述
	PayMsg string `json:"pay_msg" gorm:"type:varchar(256);default:'';comment:支付结果描述"`
}

type IPaymentLog interface {
}
