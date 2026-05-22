package biz

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
)

// CancelStatus 已取消状态
type CancelStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewCancelStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *CancelStatus {
	return &CancelStatus{order: order, repo: repo, log: log.NewHelper(logger)}
}

func (s *CancelStatus) Pay(ctx context.Context, in *Order) error {
	return errors.New("订单已取消，无法支付")
}

func (s *CancelStatus) Ship(ctx context.Context) error {
	return errors.New("订单已取消，无法发货")
}

func (s *CancelStatus) Receive(ctx context.Context) error {
	return errors.New("订单已取消，无法收货")
}

func (s *CancelStatus) Complete(ctx context.Context) error {
	return errors.New("订单已取消，无法标记完成")
}

func (s *CancelStatus) Cancel(ctx context.Context) error {
	return errors.New("订单已取消，无需重复取消")
}

func (s *CancelStatus) Refunding(ctx context.Context) error {
	return errors.New("订单已取消，无需退款")
}

func (s *CancelStatus) Refunded(ctx context.Context) error {
	return errors.New("订单已取消，无需标记已退款")
}
