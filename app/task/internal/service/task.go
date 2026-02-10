package service

import (
	"context"

	"cyi-app/app/task/internal/biz"
	"cyi-app/pkg/util/convert"

	v1 "cyi-app/api/task/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type TaskService struct {
	v1.UnimplementedTaskServer
	uc  *biz.TaskUseCase
	log log.Logger
}

func NewTaskService(uc *biz.TaskUseCase, logger log.Logger) *TaskService {
	return &TaskService{
		uc:  uc,
		log: logger,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, req *v1.CreateTaskRequest) (*v1.CreateTaskReply, error) {
	in := &biz.Task{}
	convert.Convert(req, in)
	err := s.uc.CreateTask(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.CreateTaskReply{}, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, req *v1.UpdateTaskRequest) (*v1.UpdateTaskReply, error) {
	in := &biz.Task{}
	convert.Convert(req, in)
	err := s.uc.UpdateTask(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateTaskReply{}, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, req *v1.DeleteTaskRequest) (*v1.DeleteTaskReply, error) {
	return &v1.DeleteTaskReply{}, nil
}

func (s *TaskService) GetTask(ctx context.Context, req *v1.GetTaskRequest) (*v1.GetTaskReply, error) {
	return &v1.GetTaskReply{}, nil
}

func (s *TaskService) ListTask(ctx context.Context, req *v1.ListTaskRequest) (*v1.ListTaskReply, error) {
	return &v1.ListTaskReply{}, nil
}