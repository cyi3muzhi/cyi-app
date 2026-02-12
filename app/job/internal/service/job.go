package service

import (
	"context"
	"fmt"

	"cyi-app/app/job/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

type JobService struct {
	uc  *biz.JobUseCase
	log *log.Helper
}

func NewJobService(uc *biz.JobUseCase, logger log.Logger) *JobService {
	return &JobService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// 统计用户数量
func (s *JobService) UserCount(ctx context.Context, payload *asynq.Task) error {
	// 解压处理
	payloadBytes := payload.Payload()
	fmt.Println(string(payloadBytes))
	s.log.Infof("UserCount: %s", string(payloadBytes))
	return nil
}