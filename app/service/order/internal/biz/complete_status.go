package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// CompleteStatus 已完成状态
type CompleteStatus struct {
	statusBase
}

func NewCompleteStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *CompleteStatus {
	return &CompleteStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}

// Refunding 申请退款（允许的操作）
func (s *CompleteStatus) Refunding(ctx context.Context, order *Order) error {
	s.log.WithContext(ctx).Warnf("订单开始退款, orderID: %d", order.ID)
	return s.transitionTo(ctx, order, v1.OrderStatus_Refunding)
}
