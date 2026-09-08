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

func NewPayStatus(repo OrderRepo, stockRe StockRepo, logger log.Logger) *PayStatus {
	return &PayStatus{
		statusBase: statusBase{
			repo:    repo,
			stockRe: stockRe,
			log:     log.NewHelper(logger),
		},
	}
}

// Ship 发货（允许的操作）
func (s *PayStatus) Ship(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Infof("订单发货成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Ship)
}

// Cancel 取消已支付订单：状态机先行 + 幂等库存对账
//  1. 先 CAS 抢占取消终态（Pay->Cancel）：与支付扣减/退款并发时只有一方成功
//  2. 库存对账（幂等）：逆预留方向对已扣减库存是空操作，逆扣减方向退还已扣库存，
//     两个方向由屏障裁决，支付流程中断的部分扣减也能正确对账
func (s *PayStatus) Cancel(ctx context.Context, order *Order) error {
	if err := s.transitionTo(ctx, order, v1.OrderStatus_Cancel); err != nil {
		return err
	}
	s.log.WithContext(ctx).Warnf("已支付订单取消成功, orderID: %d", order.ID)
	if err := reconcileCancelStock(ctx, s.stockRe, s.log, order); err != nil {
		s.log.WithContext(ctx).Errorf("已支付订单取消后库存对账失败（状态已取消，需补偿重试）, orderID: %d: %v", order.ID, err)
	}
	return nil
}

// Refunding 申请退款（允许的操作）
func (s *PayStatus) Refunding(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Refunding)
}
