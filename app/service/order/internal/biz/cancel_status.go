package biz

import (
	"github.com/go-kratos/kratos/v2/log"
)

// CancelStatus 已取消状态
type CancelStatus struct {
	statusBase
}

func NewCancelStatus(order OrderInterface, repo OrderRepo, logger log.Logger) *CancelStatus {
	return &CancelStatus{
		statusBase: statusBase{
			order: order,
			repo:  repo,
			log:   log.NewHelper(logger),
		},
	}
}


