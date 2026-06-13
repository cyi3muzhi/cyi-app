package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// RefundingStatus 退款中状态
type RefundingStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewRefundingStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *RefundingStatus {
	return &RefundingStatus{order: order, repo: repo, log: log.NewHelper(logger)}
}

func (s *RefundingStatus) Pay(ctx context.Context, in *Order) error {
	return NewOrderError(ErrOrderAlreadyRefunding, map[string]interface{}{
		"orderID": in.ID,
	})
}

func (s *RefundingStatus) Ship(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunding, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundingStatus) Receive(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunding, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundingStatus) Complete(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunding, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundingStatus) Cancel(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunding, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundingStatus) Refunding(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunding, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *RefundingStatus) Refunded(ctx context.Context) error {
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Infof("订单退款成功, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Refunded, s.repo, s.log.Logger())
	return nil
}
