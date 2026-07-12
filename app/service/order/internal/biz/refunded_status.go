package biz

import (
	"github.com/go-kratos/kratos/v2/log"
)

// RefundedStatus 已退款状态
type RefundedStatus struct {
	statusBase
}

func NewRefundedStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *RefundedStatus {
	return &RefundedStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}


