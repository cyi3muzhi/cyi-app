package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// ShipStatus 已发货状态
type ShipStatus struct {
	statusBase
}

func NewShipStatus(repo OrderRepo, logger log.Logger) *ShipStatus {
	return &ShipStatus{
		statusBase: statusBase{
			repo: repo,
			log:  log.NewHelper(logger),
		},
	}
}

// Receive 收货（允许的操作）
func (s *ShipStatus) Receive(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Infof("订单收货成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Complete)
}

// Refunding 申请退款（允许的操作）
func (s *ShipStatus) Refunding(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Refunding)
}
