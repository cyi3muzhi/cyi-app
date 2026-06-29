package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// CompleteStatus 已完成状态
type CompleteStatus struct {
	statusBase
}

func NewCompleteStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *CompleteStatus {
	return &CompleteStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}

func (s *CompleteStatus) Pay(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已完成，无需重复操作, orderID=%d", order.ID)
}

func (s *CompleteStatus) Ship(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已完成，无需重复操作, orderID=%d", order.ID)
}

func (s *CompleteStatus) Receive(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已完成，无需重复操作, orderID=%d", order.ID)
}

func (s *CompleteStatus) Complete(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已完成，无需重复操作, orderID=%d", order.ID)
}

func (s *CompleteStatus) Cancel(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已完成，无需重复操作, orderID=%d", order.ID)
}

func (s *CompleteStatus) Refunding(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Refunding)
}

func (s *CompleteStatus) Refunded(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已在退款中，无需重复操作, orderID=%d", order.ID)
}
