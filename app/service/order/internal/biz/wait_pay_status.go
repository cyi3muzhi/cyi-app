package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
)

// WaitPayStatus 待支付状态
type WaitPayStatus struct {
	order OrderInterface
	repo  OrderRepo
	log   *log.Helper
}

func NewWaitPayStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *WaitPayStatus {
	return &WaitPayStatus{order: order, repo: repo, log: log.NewHelper(logger)}
}

func (s *WaitPayStatus) Pay(ctx context.Context, in *Order) error {
	s.log.WithContext(ctx).Infof("订单支付成功, orderID: %d, userID: %d", in.ID, in.UserID)
	s.order.SetStatus(v1.OrderStatus_Pay, s.repo, s.log.Logger())
	in.Status = v1.OrderStatus_Pay
	if err := s.repo.UpdateOrder(ctx, in); err != nil {
		s.log.WithContext(ctx).Errorf("订单支付后更新状态失败, orderID: %d, error: %v", in.ID, err)
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
	orderID := s.order.(*Order).ID
	s.log.WithContext(ctx).Warnf("订单取消成功, orderID: %d", orderID)
	s.order.SetStatus(v1.OrderStatus_Cancel, s.repo, s.log.Logger())
	return nil
}

func (s *WaitPayStatus) Refunding(ctx context.Context) error {
	return errors.New("订单未支付，无法退款")
}

func (s *WaitPayStatus) Refunded(ctx context.Context) error {
	return errors.New("订单未退款，无法标记已退款")
}
