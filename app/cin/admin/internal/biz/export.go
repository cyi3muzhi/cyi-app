package biz

import (
	"context"
	"fmt"
	"time"

	"cyi-app/pkg/exporttask"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// ExportRepo 导出数据仓库接口（由 data 层实现）
// admin 仅负责投递任务和查询状态，不处理任务（任务处理由 task 服务消费完成）。
type ExportRepo interface {
	// EnqueueTask 投递导出任务到 Asynq 队列（由 task 服务消费）
	EnqueueTask(ctx context.Context, payload *exporttask.ExportTaskPayload) error
	// SaveTaskInfo 保存任务状态到 Redis（admin 投递时写 pending）
	SaveTaskInfo(ctx context.Context, info *exporttask.ExportTaskInfo) error
	// GetTaskInfo 查询任务状态（供前端轮询）
	GetTaskInfo(ctx context.Context, taskID string) (*exporttask.ExportTaskInfo, error)
	// GetOrGenerateDownloadURL 获取或生成下载 URL，带 Redis 缓存避免轮询时重复生成
	GetOrGenerateDownloadURL(ctx context.Context, taskID, objectKey string) (string, error)
}

// ExportUseCase 订单导出业务用例（投递端）
// admin 只负责投递任务到 asynq 队列并查询状态；
// 实际任务处理（查订单、生成 Excel、上传 OSS、更新状态）由 task 服务消费完成。
type ExportUseCase struct {
	log  *log.Helper
	repo ExportRepo
}

// NewExportUseCase 创建导出业务用例
func NewExportUseCase(repo ExportRepo, logger log.Logger) *ExportUseCase {
	return &ExportUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

// SubmitExport 投递导出任务，返回任务ID
// 流程：生成 taskID → 存 pending 状态 → 投递 Asynq 任务（由 task 服务消费）
func (uc *ExportUseCase) SubmitExport(ctx context.Context, status, dateFrom, dateTo string) (string, error) {
	taskID := uuid.NewString()
	payload := &exporttask.ExportTaskPayload{
		TaskID:   taskID,
		Status:   status,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}
	// 先持久化任务状态为 pending（24小时过期）
	info := &exporttask.ExportTaskInfo{
		TaskID:    taskID,
		Status:    exporttask.ExportStatusPending,
		CreatedAt: time.Now(),
	}
	if err := uc.repo.SaveTaskInfo(ctx, info); err != nil {
		return "", fmt.Errorf("保存任务状态失败: %w", err)
	}
	if err := uc.repo.EnqueueTask(ctx, payload); err != nil {
		return "", fmt.Errorf("投递任务失败: %w", err)
	}
	uc.log.Infof("export task submitted: task_id=%s, status=%s", taskID, status)
	return taskID, nil
}

// GetTaskStatus 查询任务状态，completed 时动态生成下载链接
func (uc *ExportUseCase) GetTaskStatus(ctx context.Context, taskID string) (*exporttask.ExportTaskInfo, string, error) {
	info, err := uc.repo.GetTaskInfo(ctx, taskID)
	if err != nil {
		return nil, "", err
	}
	// 如果任务已完成且有 objectKey，获取或生成 presigned URL（带缓存）
	var downloadURL string
	if info.Status == exporttask.ExportStatusCompleted && info.ObjectKey != "" {
		downloadURL, err = uc.repo.GetOrGenerateDownloadURL(ctx, taskID, info.ObjectKey)
		if err != nil {
			uc.log.Errorf("生成下载链接失败: task_id=%s, err=%v", taskID, err)
			// 下载链接生成失败不影响状态查询，返回空 URL
		}
	}
	return info, downloadURL, nil
}
