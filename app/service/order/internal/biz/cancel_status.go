package biz

import (
	"github.com/go-kratos/kratos/v2/log"
)

// CancelStatus 已取消状态
type CancelStatus struct {
	statusBase
}

func NewCancelStatus(repo OrderRepo, logger log.Logger) *CancelStatus {
	return &CancelStatus{
		statusBase: statusBase{
			repo: repo,
			log:  log.NewHelper(logger),
		},
	}
}


