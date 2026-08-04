package service

import (
	"context"
	"fmt"

	v1 "cyi-app/api/task/v1"
	"cyi-app/app/task/internal/biz"
	"cyi-app/pkg/exporttask"
	"cyi-app/pkg/util/convert"

	"github.com/bytedance/sonic"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// TaskService asynq 任务处理服务
// ec 处理订单导出任务（由 admin 投递，task 消费）
type TaskService struct {
	uc       *biz.TaskUseCase
	userUc   *biz.UserUseCase
	couponUc biz.CouponRepo
	ec       *biz.ExportUseCase
	log      *log.Helper
}

// NewTaskService 创建任务处理服务
func NewTaskService(uc *biz.TaskUseCase, userUc *biz.UserUseCase, couponUc biz.CouponRepo, ec *biz.ExportUseCase, logger log.Logger) *TaskService {
	return &TaskService{
		uc:       uc,
		userUc:   userUc,
		couponUc: couponUc,
		ec:       ec,
		log:      log.NewHelper(logger),
	}
}

// EmailTask 邮件发送任务（示例）
func (s *TaskService) EmailTask(ctx context.Context, payload *asynq.Task) error {
	// 解压处理
	payloadBytes := payload.Payload()
	fmt.Println(string(payloadBytes))
	s.log.Infof("CreateTask: %s", string(payloadBytes))
	return nil
}

// CreateUserTask 创建用户任务
func (s *TaskService) CreateUserTask(ctx context.Context, payload *asynq.Task) error {
	// 解压处理
	payloadBytes := payload.Payload()
	fmt.Println(string(payloadBytes))
	s.log.Infof("CreateUserTask: %s", string(payloadBytes))
	// 处理信息
	var req *v1.CreateUserTask
	if err := sonic.Unmarshal(payloadBytes, &req); err != nil {
		return err
	}
	in := &biz.User{}
	convert.Convert(req, in)
	// 调用业务逻辑
	if err := s.userUc.CreateUser(ctx, in); err != nil {
		return err
	}
	return nil
}

// HandleExportOrderTask 订单导出任务处理入口（由 admin 投递，task 消费）
// 解析 payload 后交给 ExportUseCase.ProcessExport 处理：
// 查询订单 → 生成 Excel → 上传 OSS → 更新任务状态。
func (s *TaskService) HandleExportOrderTask(ctx context.Context, t *asynq.Task) error {
	payload, err := exporttask.UnmarshalPayload(t.Payload())
	if err != nil {
		return fmt.Errorf("解析export payload失败: %w", err)
	}
	return s.ec.ProcessExport(ctx, payload)
}

// UserCount 定时统计用户数量（cron 定时任务，迁移自原 job 服务）
func (s *TaskService) UserCount(ctx context.Context, _ *asynq.Task) error {
	count, err := s.userUc.UserCount(ctx)
	if err != nil {
		return err
	}
	s.log.Infof("UserCount: %d", count)
	return nil
}

// CouponExpire 批量过期优惠券发放记录（cron 定时任务，迁移自原 job 服务）
func (s *TaskService) CouponExpire(ctx context.Context, _ *asynq.Task) error {
	affected, err := s.couponUc.BatchExpireCoupons(ctx, 100)
	if err != nil {
		s.log.Errorf("优惠券过期失败: %v", err)
		return err
	}
	s.log.Infof("CouponExpire: 已过期 %d 条优惠券记录", affected)
	return nil
}
