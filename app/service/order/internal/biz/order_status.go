package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"
)

// Order 订单接口
type OrderInterface interface {
	SetStatus(status v1.OrderStatus, repo OrderRepo)
	GetStatus() v1.OrderStatus
}

// OrderStatus 订单状态接口
type OrderStatus interface {
	// 支付
	Pay(ctx context.Context, in *Order) error
	// 发货
	Ship(ctx context.Context) error
	// 收货
	Receive(ctx context.Context) error
	// 完成
	Complete(ctx context.Context) error
	// 取消
	Cancel(ctx context.Context) error
	// 退款中
	Refunding(ctx context.Context) error
	// 已退款
	Refunded(ctx context.Context) error
}

// NewOrderStatusHandler 创建订单状态处理器
func NewOrderStatusHandler(order OrderInterface, repo OrderRepo) OrderStatus {
	switch order.GetStatus() {
	case v1.OrderStatus_WaitPay:
		return &WaitPayStatus{order: order}
	case v1.OrderStatus_Pay:
		return &PayStatus{order: order}
	case v1.OrderStatus_WaitShip:
		return &PayStatus{order: order} // 待发货状态使用已支付状态的处理逻辑
	case v1.OrderStatus_Ship:
		return &ShipStatus{order: order}
	case v1.OrderStatus_WaitReceive:
		return &ShipStatus{order: order} // 待收货状态使用已发货状态的处理逻辑
	case v1.OrderStatus_Complete:
		return &CompleteStatus{order: order}
	case v1.OrderStatus_Cancel:
		return &CancelStatus{order: order}
	case v1.OrderStatus_Refunding:
		return &RefundingStatus{order: order}
	case v1.OrderStatus_Refunded:
		return &RefundedStatus{order: order}
	default:
		return &WaitPayStatus{order: order}
	}
}
