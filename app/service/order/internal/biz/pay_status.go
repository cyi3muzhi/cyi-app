package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// PayStatus 已支付状态
type PayStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewPayStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *PayStatus {
	return &PayStatus{order: order, repo: repo, log: log.NewHelper(logger)}
}

func (s *PayStatus) Pay(ctx context.Context, in *Order) error {
	return NewOrderError(ErrOrderAlreadyPaid, map[string]interface{}{
		"orderID": in.ID,
	})
}

func (s *PayStatus) Ship(ctx context.Context) error {
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Infof("订单发货成功, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Ship, s.repo, s.log.Logger())
	return nil
}

func (s *PayStatus) Receive(ctx context.Context) error {
	return NewOrderError(ErrOrderNotShipped, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}

func (s *PayStatus) Complete(ctx context.Context) error {
	return NewOrderError(ErrOrderInvalidState, map[string]interface{}{
		"orderID":       s.order.(*Order).ID,
		"currentStatus": "pay",
		"targetStatus":  "complete",
	})
}

func (s *PayStatus) Cancel(ctx context.Context) error {
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Warnf("订单取消成功, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Cancel, s.repo, s.log.Logger())
	return nil
}

func (s *PayStatus) Refunding(ctx context.Context) error {
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Refunding, s.repo, s.log.Logger())
	return nil
}

func (s *PayStatus) Refunded(ctx context.Context) error {
	return NewOrderError(ErrOrderAlreadyRefunding, map[string]interface{}{
		"orderID": s.order.(*Order).ID,
	})
}
