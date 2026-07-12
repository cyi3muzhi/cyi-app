package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// PayStatus 已支付状态
type PayStatus struct {
	statusBase
}

func NewPayStatus(repo OrderRepo, logger log.Logger) *PayStatus {
	return &PayStatus{
		statusBase: statusBase{
			repo: repo,
			log:  log.NewHelper(logger),
		},
	}
}

// Ship 发货（允许的操作）
func (s *PayStatus) Ship(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Infof("订单发货成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Ship)
}

// Cancel 取消订单（允许的操作）
func (s *PayStatus) Cancel(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单取消成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Cancel)
}

// Refunding 申请退款（允许的操作）
func (s *PayStatus) Refunding(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Refunding)
}
