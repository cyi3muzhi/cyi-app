package data

import (
	"context"
	"cyi-app/app/task/internal/biz"
	mod "cyi-app/app/task/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type taskRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

func NewTaskRepo(data *Data, logger log.Logger) biz.TaskRepo {
	return &taskRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

func (r *taskRepo) CreateTask(ctx context.Context, in *biz.Task) error {
	Task := &mod.Task{}
	convert.Convert(in, Task)
	return nil
}

func (r *taskRepo) UpdateTask(ctx context.Context, in *biz.Task) error {
	Task := &mod.Task{}
	convert.Convert(in, Task)
	return nil
}
