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
	s.log.WithContext(ctx).Infof("订单支付状态流转, orderID: %d, userID: %d", order.ID, order.UserID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Pay)
}

// Cancel 取消订单：状态机先行 + 幂等库存对账
//  1. 先 CAS 抢占取消终态（WaitPay->Cancel）：与支付并发时只有一方成功，库存操作不再交叉
//  2. 再做库存对账（幂等）：未支付订单逆预留释放；若支付并发已扣减库存，由 pay 方向屏障自动退还
//     对账失败仅告警不回滚状态（库存操作带屏障幂等键，可由补偿任务/人工安全重跑）
func (s *WaitPayStatus) Cancel(ctx context.Context, order *Order) error {
	if err := s.transitionTo(ctx, order, v1.OrderStatus_Cancel); err != nil {
		return err
	}
	s.log.WithContext(ctx).Warnf("订单取消成功, orderID: %d", order.ID)
	if err := reconcileCancelStock(ctx, s.stockRe, s.log, order); err != nil {
		s.log.WithContext(ctx).Errorf("订单取消后库存对账失败（状态已取消，需补偿重试）, orderID: %d: %v", order.ID, err)
	}
	return nil
}
