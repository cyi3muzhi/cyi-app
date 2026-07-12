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

func NewRefundingStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *RefundingStatus {
	return &RefundingStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}

// Refunded 退款完成（允许的操作）
func (s *RefundingStatus) Refunded(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Infof("订单退款成功, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Refunded)
}
