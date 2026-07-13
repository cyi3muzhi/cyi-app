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

func NewWaitPayStatus(repo OrderRepo, stockRe StockRepo, logger log.Logger) *WaitPayStatus {
	return &WaitPayStatus{
		statusBase: statusBase{
			repo:    repo,
			stockRe: stockRe,
			log:     log.NewHelper(logger),
		},
	}
}

// Pay 支付订单（允许的操作）
func (s *WaitPayStatus) Pay(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Infof("订单支付成功, orderID: %d, userID: %d", order.ID, order.UserID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Pay)
}

// Cancel 取消订单（释放预留库存 + 状态转换）
func (s *WaitPayStatus) Cancel(ctx context.Context, order *Order) error {
	// 1. 释放预留库存（reserved-，available+）
	if err := s.releaseStockForOrder(ctx, order); err != nil {
		s.log.WithContext(ctx).Errorf("释放预留库存失败, orderID: %d, error: %v", order.ID, err)
		return err
	}

	// 2. 转换状态
	s.log.WithContext(ctx).Warnf("订单取消成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Cancel)
}
