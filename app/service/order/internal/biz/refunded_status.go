package biz

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
)

// RefundedStatus 已退款状态
type RefundedStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewRefundedStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *RefundedStatus {
	return &RefundedStatus{order: order, repo: repo, log: log.NewHelper(logger)}
}

func (s *RefundedStatus) Pay(ctx context.Context, in *Order) error {
	return errors.New("订单已退款，无法支付")
}

func (s *RefundedStatus) Ship(ctx context.Context) error {
	return errors.New("订单已退款，无法发货")
}

func (s *RefundedStatus) Receive(ctx context.Context) error {
	return errors.New("订单已退款，无法收货")
}

func (s *RefundedStatus) Complete(ctx context.Context) error {
	return errors.New("订单已退款，无法标记完成")
}

func (s *RefundedStatus) Cancel(ctx context.Context) error {
	return errors.New("订单已退款，无法取消")
}

func (s *RefundedStatus) Refunding(ctx context.Context) error {
	return errors.New("订单已退款，无需重复退款")
}

func (s *RefundedStatus) Refunded(ctx context.Context) error {
	return errors.New("订单已退款，无需重复标记")
}
