package data

import (
	"cyi-app/app/cin/admin/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// NewAsynqClient 创建 asynq 客户端（复用 admin 的 Redis 配置）
// admin 仅投递异步任务到队列，不启动 asynq.Server；
// 任务消费由 task 服务（app/task）负责。
func NewAsynqClient(cfg *conf.Data, logger log.Logger) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})
}
