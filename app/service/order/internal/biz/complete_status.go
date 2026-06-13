package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// CompleteStatus 已完成状态
type CompleteStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewCompleteStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *CompleteStatus {
	return &CompleteStatus{order: order, repo: repo, log: log.NewHelper(logger)}
}

func (s *CompleteStatus) Pay(ctx context.Context, in *Order) error {
	return NewOrderError(ErrOrderAlreadyCompleted, map[string]interface{}{
		"orderID": in.ID,
	})
}

func (s *CompleteStatus) Ship(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCompleted, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *CompleteStatus) Receive(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCompleted, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *CompleteStatus) Complete(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCompleted, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *CompleteStatus) Cancel(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyCompleted, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *CompleteStatus) Refunding(ctx context.Context) error {
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Refunding, s.repo, s.log.Logger())
	return nil
}

func (s *CompleteStatus) Refunded(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunding, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}
