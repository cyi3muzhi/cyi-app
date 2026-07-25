package service

import (
	"context"

	"cyi-app/app/job/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

type JobService struct {
	uc       *biz.JobUseCase
	userUc   biz.UserRepo
	couponUc biz.CouponRepo
	log      *log.Helper
}

func NewJobService(uc *biz.JobUseCase, userUc biz.UserRepo, couponUc biz.CouponRepo, logger log.Logger) *JobService {
	return &JobService{
		uc:       uc,
		userUc:   userUc,
		couponUc: couponUc,
		log:      log.NewHelper(logger),
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

// CouponExpire 批量过期优惠券发放记录
func (s *JobService) CouponExpire(ctx context.Context, _ *asynq.Task) error {
	affected, err := s.couponUc.BatchExpireCoupons(ctx, 100)
	if err != nil {
		s.log.Errorf("优惠券过期失败: %v", err)
		return err
	}
	s.log.Infof("CouponExpire: 已过期 %d 条优惠券记录", affected)
	return nil
}
