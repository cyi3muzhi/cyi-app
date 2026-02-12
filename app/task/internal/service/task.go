package service

import (
	"context"
	"fmt"

	"cyi-app/app/task/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

type TaskService struct {
	uc  *biz.TaskUseCase
	log *log.Helper
}

func NewTaskService(uc *biz.TaskUseCase, logger log.Logger) *TaskService {
	return &TaskService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *TaskService) EmailTask(ctx context.Context, payload *asynq.Task) error {
	// 解压处理
	payloadBytes := payload.Payload()
	fmt.Println(string(payloadBytes))
	s.log.Infof("CreateTask: %s", string(payloadBytes))
	return nil
}