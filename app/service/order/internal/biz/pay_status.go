package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"
	"errors"

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
	return errors.New("订单已支付，无需重复支付")
}

func (s *PayStatus) Ship(ctx context.Context) error {
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Infof("订单发货成功, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Ship, s.repo, s.log.Logger())
	return nil
}

func (s *PayStatus) Receive(ctx context.Context) error {
	return errors.New("订单未发货，无法收货")
}

func (s *PayStatus) Complete(ctx context.Context) error {
	return errors.New("订单未完成流程，无法标记完成")
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
	return errors.New("订单未开始退款，无法标记已退款")
}
