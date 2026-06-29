package biz

import (
	"context"

	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// CancelStatus 已取消状态
type CancelStatus struct {
	statusBase
}

func NewCancelStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *CancelStatus {
	return &CancelStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}

func (s *CancelStatus) Pay(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已取消，无法执行该操作, orderID=%d", order.ID)
}

func (s *CancelStatus) Ship(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已取消，无法执行该操作, orderID=%d", order.ID)
}

func (s *CancelStatus) Receive(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已取消，无法执行该操作, orderID=%d", order.ID)
}

func (s *CancelStatus) Complete(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已取消，无法执行该操作, orderID=%d", order.ID)
}

func (s *CancelStatus) Cancel(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已取消，无法执行该操作, orderID=%d", order.ID)
}

func (s *CancelStatus) Refunding(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已取消，无法执行该操作, orderID=%d", order.ID)
}

func (s *CancelStatus) Refunded(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已取消，无法执行该操作, orderID=%d", order.ID)
}
