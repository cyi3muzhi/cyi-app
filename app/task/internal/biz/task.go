package biz

import (
	"context"
)

type TaskRepo interface {
	// TODO: 待业务需求明确后实现
	CreateTask(ctx context.Context, in *Task) error
}

// Task 任务领域对象
// TODO: 待业务需求明确后实现完整字段
type Task struct {
}

// TaskUseCase 任务用例
// TODO: 待业务需求明确后实现完整业务逻辑
type TaskUseCase struct {
	repo TaskRepo
}

// NewTaskUseCase 创建任务用例
func NewTaskUseCase(repo TaskRepo) *TaskUseCase {
	return &TaskUseCase{repo: repo}
}

// CreateTask 创建任务
// TODO: 待业务需求明确后实现
func (uc *TaskUseCase) CreateTask(ctx context.Context, in *Task) error {
	return nil
}
