package biz

import (
	"context"
)

type JobRepo interface {
	CreateJob(ctx context.Context, in *Job) error
}


type Job struct {
}


type JobUseCase struct {
	repo JobRepo
}


func NewJobUseCase(repo JobRepo) *JobUseCase {
	return &JobUseCase{repo: repo}
}


func (uc *JobUseCase) CreateJob(ctx context.Context, in *Job) error {
	return nil
}
