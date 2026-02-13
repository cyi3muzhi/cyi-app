package service

import (
	"context"

	"cyi-app/app/job/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

type JobService struct {
	uc     *biz.JobUseCase
	userUc biz.UserRepo
	log    *log.Helper
}

func NewJobService(uc *biz.JobUseCase, userUc biz.UserRepo, logger log.Logger) *JobService {
	return &JobService{
		uc:     uc,
		userUc: userUc,
		log:    log.NewHelper(logger),
	}
}

// 统计用户数量
func (s *JobService) UserCount(ctx context.Context, _ *asynq.Task) error {
	// 统计用户数量
	count, err := s.userUc.UserCount(ctx)
	if err != nil {
		return err
	}
	s.log.Infof("UserCount: %d", count)
	return nil
}
