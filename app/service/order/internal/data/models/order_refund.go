package models

import (
	"cyi-app/pkg/util/gorm"
)

// OrderRefund 退款单（仅退款售后）
// 用户对已支付/待发货订单发起退款申请，管理员审核通过后退款并恢复库存，拒绝后订单恢复原状态
type OrderRefund struct {
	gorm.Model
	// 订单ID
	OrderID int64 `json:"order_id" gorm:"index;not null;comment:订单ID"`
	// 用户ID
	UserID int64 `json:"user_id" gorm:"index;not null;comment:用户ID"`
	// 订单号（冗余，便于列表查询展示）
	OrderNo string `json:"order_no" gorm:"type:varchar(64);not null;default:'';comment:订单号"`
	// 退款金额（单位：分，即订单实付金额）
	RefundAmount int64 `json:"refund_amount" gorm:"not null;default:0;comment:退款金额（单位：分）"`
	// 退款原因
	Reason string `json:"reason" gorm:"type:varchar(512);not null;default:'';comment:退款原因"`
	// 状态: 1-待审核 2-已通过（已退款） 3-已拒绝
	Status int32 `json:"status" gorm:"not null;default:1;comment:状态: 1-待审核 2-已通过 3-已拒绝"`
	// 申请退款前的订单状态（审核拒绝后恢复用）
	OriginStatus int32 `json:"origin_status" gorm:"not null;default:0;comment:申请前的订单状态"`
	// 拒绝原因
	RejectReason string `json:"reject_reason" gorm:"type:varchar(512);not null;default:'';comment:拒绝原因"`
	// 审核人名称
	Auditor string `json:"auditor" gorm:"type:varchar(64);not null;default:'';comment:审核人"`
	// 审核时间（Unix 秒）
	AuditTime int64 `json:"audit_time" gorm:"not null;default:0;comment:审核时间"`
}

type IOrderRefund interface {
}
