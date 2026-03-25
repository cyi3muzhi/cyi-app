package biz

import (
	"context"
	"errors"
	v1 "cyi-app/api/service/order/v1"
	"fmt"
)

// ShipStatus 已发货状态
type ShipStatus struct {
	order OrderInterface
	repo OrderRepo
}

func NewShipStatus(order OrderInterface, repo OrderRepo) *ShipStatus {
	return &ShipStatus{order: order, repo: repo}
}



func (s *ShipStatus) Pay(ctx context.Context, in *Order) error {
	return errors.New("订单已支付，无需重复支付")
}

func (s *ShipStatus) Ship(ctx context.Context) error {
	return errors.New("订单已发货，无需重复发货")
}

func (s *ShipStatus) Receive(ctx context.Context) error {
	fmt.Println("订单收货成功")
	s.order.SetStatus(v1.OrderStatus_Complete, s.repo)
	return nil
}

func (s *ShipStatus) Complete(ctx context.Context) error {
	return errors.New("订单未收货，无法标记完成")
}

func (s *ShipStatus) Cancel(ctx context.Context) error {
	return errors.New("订单已发货，无法取消")
}

func (s *ShipStatus) Refunding(ctx context.Context) error {
	fmt.Println("订单开始退款")
	s.order.SetStatus(v1.OrderStatus_Refunding, s.repo)
	return nil
}

func (s *ShipStatus) Refunded(ctx context.Context) error {
	return errors.New("订单未开始退款，无法标记已退款")
}
