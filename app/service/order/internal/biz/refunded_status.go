package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// RefundedStatus 已退款状态
type RefundedStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewRefundedStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *RefundedStatus {
	return &RefundedStatus{order: order, repo: repo, log: log.NewHelper(logger)}
}

func (s *RefundedStatus) Pay(ctx context.Context, in *Order) error {
	return NewOrderError(ErrOrderAlreadyRefunded, map[string]interface{}{
		"orderID": in.ID,
	})
}

func (s *RefundedStatus) Ship(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunded, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundedStatus) Receive(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunded, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundedStatus) Complete(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunded, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundedStatus) Cancel(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunded, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundedStatus) Refunding(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunded, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundedStatus) Refunded(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunded, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}
