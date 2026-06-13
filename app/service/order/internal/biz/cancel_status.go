package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// CancelStatus 已取消状态
type CancelStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewCancelStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *CancelStatus {
	return &CancelStatus{order: order, repo: repo, log: log.NewHelper(logger)}
}

func (s *CancelStatus) Pay(ctx context.Context, in *Order) error {
	return NewOrderError(ErrOrderAlreadyCancelled, map[string]interface{}{
		"orderID": in.ID,
	})
}

func (s *CancelStatus) Ship(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCancelled, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *CancelStatus) Receive(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCancelled, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *CancelStatus) Complete(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCancelled, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *CancelStatus) Cancel(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCancelled, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *CancelStatus) Refunding(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCancelled, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *CancelStatus) Refunded(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCancelled, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}
