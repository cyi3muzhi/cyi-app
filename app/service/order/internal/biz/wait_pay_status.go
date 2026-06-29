package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// WaitPayStatus 待支付状态
type WaitPayStatus struct {
	statusBase
}

func NewWaitPayStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *WaitPayStatus {
	return &WaitPayStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}

func (s *WaitPayStatus) Pay(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Infof("订单支付成功, orderID: %d, userID: %d", order.ID, order.UserID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Pay)
}

func (s *WaitPayStatus) Ship(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单未支付，无法执行该操作, orderID=%d", order.ID)
}

func (s *WaitPayStatus) Receive(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单未发货，无法执行该操作, orderID=%d", order.ID)
}

func (s *WaitPayStatus) Complete(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d, currentStatus=wait_pay", order.ID)
}

func (s *WaitPayStatus) Cancel(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单取消成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Cancel)
}

func (s *WaitPayStatus) Refunding(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单未支付，无法执行该操作, orderID=%d", order.ID)
}

func (s *WaitPayStatus) Refunded(ctx context.Context, order *Order) error {
	return v1.ErrorOrderStatusInvalid("订单已退款，无需重复操作, orderID=%d", order.ID)
}
