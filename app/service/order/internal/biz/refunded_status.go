package biz

import (
	"context"

	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// RefundedStatus 已退款状态
type RefundedStatus struct {
	statusBase
}

func NewRefundedStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *RefundedStatus {
	return &RefundedStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}

func (s *RefundedStatus) Pay(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已退款，无需重复操作, orderID=%d", order.ID)
}

func (s *RefundedStatus) Ship(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已退款，无需重复操作, orderID=%d", order.ID)
}

func (s *RefundedStatus) Receive(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已退款，无需重复操作, orderID=%d", order.ID)
}

func (s *RefundedStatus) Complete(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已退款，无需重复操作, orderID=%d", order.ID)
}

func (s *RefundedStatus) Cancel(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已退款，无需重复操作, orderID=%d", order.ID)
}

func (s *RefundedStatus) Refunding(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已退款，无需重复操作, orderID=%d", order.ID)
}

func (s *RefundedStatus) Refunded(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已退款，无需重复操作, orderID=%d", order.ID)
}
