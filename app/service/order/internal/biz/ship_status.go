package biz

import (
	"context"
	"errors"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// ShipStatus 已发货状态
type ShipStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewShipStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *ShipStatus {
	return &ShipStatus{order: order, repo: repo, log: log.NewHelper(logger)}
}



func (s *ShipStatus) Pay(ctx context.Context, in *Order) error {
	return errors.New("订单已支付，无需重复支付")
}

func (s *ShipStatus) Ship(ctx context.Context) error {
	return errors.New("订单已发货，无需重复发货")
}

func (s *ShipStatus) Receive(ctx context.Context) error {
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Infof("订单收货成功, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Complete, s.repo, s.log.Logger())
	return nil
}

func (s *ShipStatus) Complete(ctx context.Context) error {
	return errors.New("订单未收货，无法标记完成")
}

func (s *ShipStatus) Cancel(ctx context.Context) error {
	return errors.New("订单已发货，无法取消")
}

func (s *ShipStatus) Refunding(ctx context.Context) error {
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Refunding, s.repo, s.log.Logger())
	return nil
}

func (s *ShipStatus) Refunded(ctx context.Context) error {
	return errors.New("订单未开始退款，无法标记已退款")
}
