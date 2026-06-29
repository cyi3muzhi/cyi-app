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

// statusBase 状态处理器基础结构，提供统一的状态转换方法
type statusBase struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

// transitionTo 统一的状态转换方法，更新内存状态并持久化到数据库
func (b *statusBase) transitionTo(ctx context.Context, order *Order, newStatus v1.OrderStatus) error {
	order.Status = newStatus
	b.order.SetStatus(newStatus, b.repo, b.log.Logger())
	if err := b.repo.UpdateOrder(ctx, order); err != nil {
		b.log.WithContext(ctx).Errorf("订单状态更新失败, orderID: %d, newStatus: %v, error: %v", order.ID, newStatus, err)
		return err
	}
	return nil
}

// OrderStatus 订单状态接口，所有方法统一接收 order 参数，消除类型断言依赖
type OrderStatus interface {
	// 支付
	Pay(ctx context.Context, order *Order) error
	// 发货
	Ship(ctx context.Context, order *Order) error
	// 收货
	Receive(ctx context.Context, order *Order) error
	// 完成
	Complete(ctx context.Context, order *Order) error
	// 取消
	Cancel(ctx context.Context, order *Order) error
	// 退款中
	Refunding(ctx context.Context, order *Order) error
	// 已退款
	Refunded(ctx context.Context, order *Order) error
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

// Pay 支付订单
func (o *Order) Pay(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Pay(ctx, o)
}

// Ship 发货
func (o *Order) Ship(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Ship(ctx, o)
}

// Receive 收货
func (o *Order) Receive(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Receive(ctx, o)
}

// Complete 完成
func (o *Order) Complete(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Complete(ctx, o)
}

// Cancel 取消
func (o *Order) Cancel(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Cancel(ctx, o)
}

// Refunding 退款中
func (o *Order) Refunding(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Refunding(ctx, o)
}

// Refunded 已退款
func (o *Order) Refunded(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Refunded(ctx, o)
}
