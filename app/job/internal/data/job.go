package data

import (
	"context"
	"cyi-app/app/job/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type jobRepo struct {
	data  *Data
	cache redis.UniversalClient
	log   *log.Helper
}

func NewJobRepo(data *Data, logger log.Logger) biz.JobRepo {
	return &jobRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

func (r *jobRepo) CreateJob(ctx context.Context, in *biz.Job) error {
	return nil
}
