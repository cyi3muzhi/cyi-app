package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// ShipStatus 已发货状态
type ShipStatus struct {
	statusBase
}

func NewShipStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *ShipStatus {
	return &ShipStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}

func (s *ShipStatus) Pay(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已支付，无需重复支付, orderID=%d", order.ID)
}

func (s *ShipStatus) Ship(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已发货，无需重复发货, orderID=%d", order.ID)
}

func (s *ShipStatus) Receive(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Infof("订单收货成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Complete)
}

func (s *ShipStatus) Complete(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d, currentStatus=ship", order.ID)
}

func (s *ShipStatus) Cancel(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已发货，无法执行该操作, orderID=%d", order.ID)
}

func (s *ShipStatus) Refunding(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Refunding)
}

func (s *ShipStatus) Refunded(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已在退款中，无需重复操作, orderID=%d", order.ID)
}
