package biz

import (
	"context"
)

type TaskRepo interface {
	CreateTask(ctx context.Context, in *Task) error
}


type Task struct {
}


type TaskUseCase struct {
	repo TaskRepo
}


func NewTaskUseCase(repo TaskRepo) *TaskUseCase {
	return &TaskUseCase{repo: repo}
}


func (uc *TaskUseCase) CreateTask(ctx context.Context, in *Task) error {
	return nil
}
