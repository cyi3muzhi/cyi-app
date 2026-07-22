package biz

import (
	"github.com/go-kratos/kratos/v2/log"
)

// FinalStatus 终端状态处理器（已取消/已退款等不可再变更的状态）
// 所有方法由 statusBase 的默认拒绝实现覆盖，无需自定义行为
type FinalStatus struct {
	statusBase
}

// NewFinalStatus 创建终端状态处理器
func NewFinalStatus(repo OrderRepo, stockRe StockRepo, logger log.Logger) *FinalStatus {
	return &FinalStatus{
		statusBase: statusBase{
			repo:    repo,
			stockRe: stockRe,
			log:     log.NewHelper(logger),
		},
	}
}
