package biz

import (
	"github.com/go-kratos/kratos/v2/log"
)

// CancelStatus 已取消状态
type CancelStatus struct {
	statusBase
}

func NewCancelStatus(repo OrderRepo, stockRe StockRepo, logger log.Logger) *CancelStatus {
	return &CancelStatus{
		statusBase: statusBase{
			repo:    repo,
			stockRe: stockRe,
			log:     log.NewHelper(logger),
		},
	}
}


