package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// RefundingStatus 退款中状态
type RefundingStatus struct {
	statusBase
}

func NewRefundingStatus(repo OrderRepo, stockRe StockRepo, logger log.Logger) *RefundingStatus {
	return &RefundingStatus{
		statusBase: statusBase{
			repo:    repo,
			stockRe: stockRe,
			log:     log.NewHelper(logger),
		},
	}
}

// Refunded 退款审核通过：状态机先行 + 幂等库存对账
//  1. 先 CAS 流转退款终态（Refunding->Refunded）
//  2. 库存对账（幂等）：逆扣减方向退还已支付扣减的库存（total+，available+），逆预留方向为空操作；
//     对账失败仅告警（屏障幂等保证可安全重跑），不回滚退款状态
func (s *RefundingStatus) Refunded(ctx context.Context, order *Order) error {
	if err := s.transitionTo(ctx, order, v1.OrderStatus_Refunded); err != nil {
		return err
	}
	s.log.WithContext(ctx).Infof("订单退款成功, orderID: %d", order.ID)
	if err := reconcileCancelStock(ctx, s.stockRe, s.log, order); err != nil {
		s.log.WithContext(ctx).Errorf("退款成功后库存对账失败（状态已退款，需补偿重试）, orderID: %d: %v", order.ID, err)
	}
	return nil
}
