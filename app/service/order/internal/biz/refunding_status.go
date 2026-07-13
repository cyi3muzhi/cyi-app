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

// Refunded 退款完成（增加库存 + 状态转换）
func (s *RefundingStatus) Refunded(ctx context.Context, order *Order) error {
	// 1. 增加库存（恢复已扣减的库存 total+，available+）
	if err := s.addStockForOrder(ctx, order); err != nil {
		s.log.WithContext(ctx).Errorf("增加库存失败, orderID: %d, error: %v", order.ID, err)
		return err
	}

	// 2. 转换状态
	s.log.WithContext(ctx).Infof("订单退款成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Refunded)
}
