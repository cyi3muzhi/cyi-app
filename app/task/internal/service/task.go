package service

import (
	"context"
	"fmt"

	v1 "cyi-app/api/task/v1"
	"cyi-app/app/task/internal/biz"
	"cyi-app/pkg/util/convert"

	"github.com/bytedance/sonic"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

type TaskService struct {
	uc     *biz.TaskUseCase
	userUc *biz.UserUseCase
	log    *log.Helper
}

func NewTaskService(uc *biz.TaskUseCase, userUc *biz.UserUseCase, logger log.Logger) *TaskService {
	return &TaskService{
		uc:     uc,
		userUc: userUc,
		log:    log.NewHelper(logger),
	}
}

func (s *TaskService) EmailTask(ctx context.Context, payload *asynq.Task) error {
	// 解压处理
	payloadBytes := payload.Payload()
	fmt.Println(string(payloadBytes))
	s.log.Infof("CreateTask: %s", string(payloadBytes))
	return nil
}

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
