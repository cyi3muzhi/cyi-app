package biz

import "context"

// OrderTimeoutTaskRepo 订单超时任务投递仓库接口（由 data 层基于 asynq 延迟队列实现）
// 下单成功后投递延迟任务，超时后由 task 服务消费并自动取消未支付订单。
type OrderTimeoutTaskRepo interface {
	// EnqueueTimeoutCancel 投递订单超时取消延迟任务
	EnqueueTimeoutCancel(ctx context.Context, orderID int64) error
}
