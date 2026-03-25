package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"
	"errors"
	"fmt"
)

// RefundingStatus 退款中状态
type RefundingStatus struct {
	order OrderInterface
	repo OrderRepo
}

func NewRefundingStatus(order OrderInterface, repo OrderRepo) *RefundingStatus {
	return &RefundingStatus{order: order, repo: repo}
}

func (s *RefundingStatus) Pay(ctx context.Context, in *Order) error {
	return errors.New("订单退款中，无法支付")
}

func (s *RefundingStatus) Ship(ctx context.Context) error {
	return errors.New("订单退款中，无法发货")
}

func (s *RefundingStatus) Receive(ctx context.Context) error {
	return errors.New("订单退款中，无法收货")
}

func (s *RefundingStatus) Complete(ctx context.Context) error {
	return errors.New("订单退款中，无法标记完成")
}

func (s *RefundingStatus) Cancel(ctx context.Context) error {
	return errors.New("订单退款中，无法取消")
}

func (s *RefundingStatus) Refunding(ctx context.Context) error {
	return errors.New("订单已在退款中，无需重复操作")
}

func (s *RefundingStatus) Refunded(ctx context.Context) error {
	fmt.Println("订单退款成功")
	s.order.SetStatus(v1.OrderStatus_Refunded, s.repo)
	return nil
}
