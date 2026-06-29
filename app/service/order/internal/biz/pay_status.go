package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// PayStatus 已支付状态
type PayStatus struct {
	statusBase
}

func NewPayStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *PayStatus {
	return &PayStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}

func (s *PayStatus) Pay(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已支付，无需重复支付, orderID=%d", order.ID)
}

func (s *PayStatus) Ship(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Infof("订单发货成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Ship)
}

func (s *PayStatus) Receive(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单未发货，无法执行该操作, orderID=%d", order.ID)
}

func (s *PayStatus) Complete(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d, currentStatus=pay", order.ID)
}

func (s *PayStatus) Cancel(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单取消成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Cancel)
}

func (s *PayStatus) Refunding(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Refunding)
}

func (s *PayStatus) Refunded(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已在退款中，无需重复操作, orderID=%d", order.ID)
}
