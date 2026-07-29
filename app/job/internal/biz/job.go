package biz

import (
	"context"
)

type JobRepo interface {
	// TODO: 待业务需求明确后实现
	CreateJob(ctx context.Context, in *Job) error
}

// Job 任务领域对象
// TODO: 待业务需求明确后实现完整字段
type Job struct {
}

// JobUseCase 任务用例
// TODO: 待业务需求明确后实现完整业务逻辑
type JobUseCase struct {
	repo JobRepo
}

// NewJobUseCase 创建任务用例
func NewJobUseCase(repo JobRepo) *JobUseCase {
	return &JobUseCase{repo: repo}
}

// CreateJob 创建任务
// TODO: 待业务需求明确后实现
func (uc *JobUseCase) CreateJob(ctx context.Context, in *Job) error {
	return nil
}
