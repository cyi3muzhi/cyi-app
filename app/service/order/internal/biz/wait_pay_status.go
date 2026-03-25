package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"
	"errors"
	"fmt"
)

// WaitPayStatus 待支付状态
type WaitPayStatus struct {
	order OrderInterface
	repo OrderRepo
}
func NewWaitPayStatus(order OrderInterface, repo OrderRepo) *WaitPayStatus {
	return &WaitPayStatus{order: order, repo: repo}
}

func (s *WaitPayStatus) Pay(ctx context.Context, in *Order) error {
	fmt.Println("订单支付成功")
	s.order.SetStatus(v1.OrderStatus_Pay, s.repo)
	in.Status = v1.OrderStatus_Pay
	if err := s.repo.UpdateOrder(ctx, in); err != nil {
		return err
	}
	return nil
}

func (s *WaitPayStatus) Ship(ctx context.Context) error {
	return errors.New("订单未支付，无法发货")
}

func (s *WaitPayStatus) Receive(ctx context.Context) error {
	return errors.New("订单未发货，无法收货")
}

func (s *WaitPayStatus) Complete(ctx context.Context) error {
	return errors.New("订单未完成流程，无法标记完成")
}

func (s *WaitPayStatus) Cancel(ctx context.Context) error {
	fmt.Println("订单取消成功")
	s.order.SetStatus(v1.OrderStatus_Cancel, s.repo)
	return nil
}

func (s *WaitPayStatus) Refunding(ctx context.Context) error {
	return errors.New("订单未支付，无法退款")
}

func (s *WaitPayStatus) Refunded(ctx context.Context) error {
	return errors.New("订单未退款，无法标记已退款")
}
