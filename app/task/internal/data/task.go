package data

import (
	"context"
	"cyi-app/app/task/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type taskRepo struct {
	data  *Data
	cache redis.UniversalClient
	log   *log.Helper
}

func NewTaskRepo(data *Data, logger log.Logger) biz.TaskRepo {
	return &taskRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

func (r *taskRepo) CreateTask(ctx context.Context, in *biz.Task) error {
	return nil
}
