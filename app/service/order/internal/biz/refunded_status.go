package biz

import (
	"github.com/go-kratos/kratos/v2/log"
)

// RefundedStatus 已退款状态
type RefundedStatus struct {
	statusBase
}

func NewRefundedStatus(repo OrderRepo, stockRe StockRepo, logger log.Logger) *RefundedStatus {
	return &RefundedStatus{
		statusBase: statusBase{
			repo:    repo,
			stockRe: stockRe,
			log:     log.NewHelper(logger),
		},
	}
}


