package biz

import (
	"context"
	"io"
	"testing"

	"cyi-app/pkg/exporttask"

	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// mockExportRepo 模拟 ExportRepo 接口
type mockExportRepo struct {
	enqueueTaskFn       func(ctx context.Context, payload *exporttask.ExportTaskPayload) error
	saveTaskInfoFn      func(ctx context.Context, info *exporttask.ExportTaskInfo) error
	getTaskInfoFn       func(ctx context.Context, taskID string) (*exporttask.ExportTaskInfo, error)
	getOrGenerateURLFn func(ctx context.Context, taskID, objectKey string) (string, error)
}

func (m *mockExportRepo) EnqueueTask(ctx context.Context, payload *exporttask.ExportTaskPayload) error {
	return m.enqueueTaskFn(ctx, payload)
}

func (m *mockExportRepo) SaveTaskInfo(ctx context.Context, info *exporttask.ExportTaskInfo) error {
	return m.saveTaskInfoFn(ctx, info)
}

func (m *mockExportRepo) GetTaskInfo(ctx context.Context, taskID string) (*exporttask.ExportTaskInfo, error) {
	return m.getTaskInfoFn(ctx, taskID)
}

func (m *mockExportRepo) GetOrGenerateDownloadURL(ctx context.Context, taskID, objectKey string) (string, error) {
	return m.getOrGenerateURLFn(ctx, taskID, objectKey)
}

// TestExportUseCase_SubmitExport_Success 测试提交导出任务成功：状态为 pending 且 taskID 不为空
func TestExportUseCase_SubmitExport_Success(t *testing.T) {
	var savedInfo *exporttask.ExportTaskInfo
	var enqueuedPayload *exporttask.ExportTaskPayload

	uc := NewExportUseCase(&mockExportRepo{
		saveTaskInfoFn: func(_ context.Context, info *exporttask.ExportTaskInfo) error {
			savedInfo = info
			return nil
		},
		enqueueTaskFn: func(_ context.Context, payload *exporttask.ExportTaskPayload) error {
			enqueuedPayload = payload
			return nil
		},
	}, kratoslog.NewStdLogger(io.Discard))

	taskID, err := uc.SubmitExport(context.Background(), "paid", "2024-01-01", "2024-01-31")
	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)

	// 验证 SaveTaskInfo 被调用，且状态为 pending
	assert.NotNil(t, savedInfo)
	assert.Equal(t, taskID, savedInfo.TaskID)
	assert.Equal(t, exporttask.ExportStatusPending, savedInfo.Status)
	assert.NotNil(t, savedInfo.CreatedAt)

	// 验证 EnqueueTask 被调用，且 payload 正确
	assert.NotNil(t, enqueuedPayload)
	assert.Equal(t, taskID, enqueuedPayload.TaskID)
	assert.Equal(t, "paid", enqueuedPayload.Status)
	assert.Equal(t, "2024-01-01", enqueuedPayload.DateFrom)
	assert.Equal(t, "2024-01-31", enqueuedPayload.DateTo)
}

// TestExportUseCase_SubmitExport_SaveTaskInfoError 测试保存任务状态失败时错误传播
func TestExportUseCase_SubmitExport_SaveTaskInfoError(t *testing.T) {
	expectedErr := io.ErrUnexpectedEOF
	uc := NewExportUseCase(&mockExportRepo{
		saveTaskInfoFn: func(_ context.Context, info *exporttask.ExportTaskInfo) error {
			return expectedErr
		},
	}, kratoslog.NewStdLogger(io.Discard))

	taskID, err := uc.SubmitExport(context.Background(), "all", "", "")
	assert.ErrorIs(t, err, expectedErr)
	assert.Empty(t, taskID)
}

// TestExportUseCase_SubmitExport_EnqueueTaskError 测试保存状态成功但投递任务失败时错误传播
// 验证即使 EnqueueTask 失败，SaveTaskInfo 也已被调用（有错误的副作用）
func TestExportUseCase_SubmitExport_EnqueueTaskError(t *testing.T) {
	var savedInfo *exporttask.ExportTaskInfo
	expectedErr := io.ErrClosedPipe

	uc := NewExportUseCase(&mockExportRepo{
		saveTaskInfoFn: func(_ context.Context, info *exporttask.ExportTaskInfo) error {
			savedInfo = info
			return nil
		},
		enqueueTaskFn: func(_ context.Context, payload *exporttask.ExportTaskPayload) error {
			return expectedErr
		},
	}, kratoslog.NewStdLogger(io.Discard))

	taskID, err := uc.SubmitExport(context.Background(), "pending", "2024-06-01", "2024-06-30")
	assert.ErrorIs(t, err, expectedErr)
	assert.Empty(t, taskID)

	// 验证 SaveTaskInfo 已被调用（即使后续 EnqueueTask 失败）
	assert.NotNil(t, savedInfo)
	assert.Equal(t, exporttask.ExportStatusPending, savedInfo.Status)
}

// TestExportUseCase_SubmitExport_TaskIDUniqueness 测试连续两次提交返回不同的 taskID
func TestExportUseCase_SubmitExport_TaskIDUniqueness(t *testing.T) {
	var callCount int
	uc := NewExportUseCase(&mockExportRepo{
		saveTaskInfoFn: func(_ context.Context, info *exporttask.ExportTaskInfo) error {
			callCount++
			return nil
		},
		enqueueTaskFn: func(_ context.Context, payload *exporttask.ExportTaskPayload) error {
			return nil
		},
	}, kratoslog.NewStdLogger(io.Discard))

	taskID1, err := uc.SubmitExport(context.Background(), "all", "", "")
	assert.NoError(t, err)

	taskID2, err := uc.SubmitExport(context.Background(), "all", "", "")
	assert.NoError(t, err)

	// 两次生成的 taskID 必须不同
	assert.NotEqual(t, taskID1, taskID2)
	assert.Equal(t, 2, callCount)
}
