package biz

import (
	"context"
	"errors"
	v1 "cyi-app/api/service/order/v1"
	"fmt"
)

// CompleteStatus 已完成状态
type CompleteStatus struct {
	order OrderInterface
	repo OrderRepo
}

func NewCompleteStatus(order OrderInterface, repo OrderRepo) *CompleteStatus {
	return &CompleteStatus{order: order, repo: repo}
}

func (s *CompleteStatus) Pay(ctx context.Context, in *Order) error {
	return errors.New("订单已完成，无需支付")
}

func (s *CompleteStatus) Ship(ctx context.Context) error {
	return errors.New("订单已完成，无需发货")
}

func (s *CompleteStatus) Receive(ctx context.Context) error {
	return errors.New("订单已完成，无需收货")
}

func (s *CompleteStatus) Complete(ctx context.Context) error {
	return errors.New("订单已完成，无需重复标记")
}

func (s *CompleteStatus) Cancel(ctx context.Context) error {
	return errors.New("订单已完成，无法取消")
}

func (s *CompleteStatus) Refunding(ctx context.Context) error {
	fmt.Println("订单开始退款")
	s.order.SetStatus(v1.OrderStatus_Refunding, s.repo)
	return nil
}

func (s *CompleteStatus) Refunded(ctx context.Context) error {
	return errors.New("订单未开始退款，无法标记已退款")
}
