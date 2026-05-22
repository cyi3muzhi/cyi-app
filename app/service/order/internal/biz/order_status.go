package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// Order 订单接口
type OrderInterface interface {
	SetStatus(status v1.OrderStatus, repo OrderRepo, logger log.Logger)
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
func NewOrderStatusHandler(order OrderInterface, repo OrderRepo, logger log.Logger) OrderStatus {
	switch order.GetStatus() {
	case v1.OrderStatus_WaitPay:
		return NewWaitPayStatus(order, repo, logger)
	case v1.OrderStatus_Pay:
		return NewPayStatus(order, repo, logger)
	case v1.OrderStatus_WaitShip:
		return NewPayStatus(order, repo, logger) // 待发货状态使用已支付状态的处理逻辑
	case v1.OrderStatus_Ship:
		return NewShipStatus(order, repo, logger)
	case v1.OrderStatus_WaitReceive:
		return NewShipStatus(order, repo, logger) // 待收货状态使用已发货状态的处理逻辑
	case v1.OrderStatus_Complete:
		return NewCompleteStatus(order, repo, logger)
	case v1.OrderStatus_Cancel:
		return NewCancelStatus(order, repo, logger)
	case v1.OrderStatus_Refunding:
		return NewRefundingStatus(order, repo, logger)
	case v1.OrderStatus_Refunded:
		return NewRefundedStatus(order, repo, logger)
	default:
		return NewWaitPayStatus(order, repo, logger)
	}
}


// GetStatus 获取订单状态
func (o *Order) GetStatus() v1.OrderStatus {
	return o.Status
}

// SetStatus 设置状态并更新状态处理器
func (o *Order) SetStatus(newStatus v1.OrderStatus, repo OrderRepo, logger log.Logger) {
	o.Status = newStatus
	o.statusHandler = NewOrderStatusHandler(o, repo, logger)
}

// 支付
func (o *Order) Pay(ctx context.Context, in *Order) error {
	return o.statusHandler.Pay(ctx, in)
}

// 发货
func (o *Order) Ship(ctx context.Context) error {
	return o.statusHandler.Ship(ctx)
}

// 收货
func (o *Order) Receive(ctx context.Context) error {
	return o.statusHandler.Receive(ctx)
}

// 完成
func (o *Order) Complete(ctx context.Context) error {
	return o.statusHandler.Complete(ctx)
}

// 取消
func (o *Order) Cancel(ctx context.Context) error {
	return o.statusHandler.Cancel(ctx)
}

// 退款中
func (o *Order) Refunding(ctx context.Context) error {
	return o.statusHandler.Refunding(ctx)
}

// 已退款
func (o *Order) Refunded(ctx context.Context) error {
	return o.statusHandler.Refunded(ctx)
}
