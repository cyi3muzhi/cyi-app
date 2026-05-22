package biz

import (
	"context"
	"errors"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// CompleteStatus 已完成状态
type CompleteStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewCompleteStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *CompleteStatus {
	return &CompleteStatus{order: order, repo: repo, log: log.NewHelper(logger)}
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
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Refunding, s.repo, s.log.Logger())
	return nil
}

func (s *CompleteStatus) Refunded(ctx context.Context) error {
	return errors.New("订单未开始退款，无法标记已退款")
}
