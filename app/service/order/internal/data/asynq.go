package data

import (
	"context"
	"fmt"

	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/conf"
	"cyi-app/pkg/ordertask"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// NewAsynqClient 创建 asynq 客户端（复用订单服务的 Redis 配置）
// 订单服务仅投递延迟任务到队列，不启动 asynq.Server；
// 任务消费由 task 服务（app/task）负责。
func NewAsynqClient(cfg *conf.Data, logger log.Logger) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})
}

// orderTimeoutTaskRepo 订单超时取消任务的投递仓库实现（基于 asynq 延迟队列）
type orderTimeoutTaskRepo struct {
	log    *log.Helper
	client *asynq.Client
}

// NewOrderTimeoutTaskRepo 创建订单超时任务投递仓库
func NewOrderTimeoutTaskRepo(client *asynq.Client, logger log.Logger) biz.OrderTimeoutTaskRepo {
	return &orderTimeoutTaskRepo{
		log:    log.NewHelper(logger),
		client: client,
	}
}

// EnqueueTimeoutCancel 投递订单超时取消延迟任务
// 30 分钟后由 task 服务消费：若订单仍为待付款则自动取消，否则跳过。
// 使用固定 TaskID 保证同一订单仅存在一个待执行的超时任务。
func (r *orderTimeoutTaskRepo) EnqueueTimeoutCancel(ctx context.Context, orderID int64) error {
	payload, err := ordertask.MarshalPayload(&ordertask.OrderTimeoutPayload{OrderID: orderID})
	if err != nil {
		return fmt.Errorf("序列化订单超时任务payload失败: %w", err)
	}
	task := asynq.NewTask(ordertask.OrderTimeoutTaskType, payload,
		asynq.ProcessIn(ordertask.OrderTimeoutDelay), // 30 分钟后执行
		asynq.TaskID(ordertask.TaskID(orderID)),      // 同一订单只保留一个待执行任务
		asynq.MaxRetry(2),                            // 失败最多重试 2 次
	)
	info, err := r.client.EnqueueContext(ctx, task)
	if err != nil {
		// 同一订单已存在待执行的超时任务（TaskID 冲突）时忽略，避免影响下单主流程
		if err == asynq.ErrTaskIDConflict {
			r.log.Warnf("订单 %d 已存在超时任务，跳过重复投递", orderID)
			return nil
		}
		return fmt.Errorf("投递订单超时任务失败: %w", err)
	}
	r.log.Infof("订单超时任务已投递: order_id=%d, asynq_id=%s", orderID, info.ID)
	return nil
}
