package biz

import (
	"context"
	"fmt"
	"time"

	"cyi-app/pkg/exporttask"

	"github.com/go-kratos/kratos/v2/log"
)

// ExportOrder 订单导出所需的订单字段（精简版，仅 Excel 列使用）
// 由 data 层通过 order gRPC + user gRPC 查询并组装。
type ExportOrder struct {
	OrderNo    string             // 订单号
	CreatedAt   string             // 下单时间（格式化字符串）
	UserName    string             // 用户名
	Amount      int64              // 订单金额（单位：分）
	Payment     string             // 支付方式
	StatusText  string             // 状态文本
	Products    []*ExportOrderProduct
}

// ExportOrderProduct 订单商品明细（Excel"商品信息"列使用）
type ExportOrderProduct struct {
	Name string
	Spec string
	Qty  int32
}

// OrderQueryRepo 订单查询仓库接口（由 data 层通过 order/user gRPC 实现）
// task 服务作为消费端，自行查询订单数据，不依赖 admin 服务。
type OrderQueryRepo interface {
	// ListOrdersForExport 分页查询订单用于导出
	// status 为管理端状态编码（如 paid/shipped），dateFrom/dateTo 为 "2006-01-02"
	ListOrdersForExport(ctx context.Context, page, pageSize int32, status, dateFrom, dateTo string) ([]*ExportOrder, int32, error)
}

// ExportRepo 导出任务状态仓库接口（由 data 层基于 Redis 实现）
// admin 投递时写 pending，task 消费时写 processing/completed/failed。
type ExportRepo interface {
	// SaveTaskInfo 保存任务状态到 Redis（覆盖写）
	SaveTaskInfo(ctx context.Context, info *exporttask.ExportTaskInfo) error
	// GetTaskInfo 查询任务状态
	GetTaskInfo(ctx context.Context, taskID string) (*exporttask.ExportTaskInfo, error)
	// UpdateTaskProgress 更新任务进度（处理中状态）
	UpdateTaskProgress(ctx context.Context, taskID string, processed, total int32) error
}

// ExcelWriter Excel 生成器接口（由 data 层实现，生成并上传到 OSS）
type ExcelWriter interface {
	// GenerateAndUpload 生成订单 Excel 并上传到 OSS，返回 objectKey
	GenerateAndUpload(ctx context.Context, orders []*ExportOrder) (objectKey string, err error)
}

// ExportUseCase 订单导出业务用例（消费端）
// 由 asynq handler 调用 ProcessExport 处理导出任务。
type ExportUseCase struct {
	log         *log.Helper
	repo        ExportRepo
	excelWriter ExcelWriter
	orderRepo   OrderQueryRepo
}

// NewExportUseCase 创建导出业务用例
func NewExportUseCase(repo ExportRepo, ew ExcelWriter, orderRepo OrderQueryRepo, logger log.Logger) *ExportUseCase {
	return &ExportUseCase{
		log:         log.NewHelper(logger),
		repo:        repo,
		excelWriter: ew,
		orderRepo:   orderRepo,
	}
}

// ProcessExport 处理导出任务（由 asynq handler 调用）
// 流程：标记 processing → 分页拉取订单 → 生成 Excel → 上传 OSS → 更新 completed/failed
func (uc *ExportUseCase) ProcessExport(ctx context.Context, payload *exporttask.ExportTaskPayload) error {
	taskID := payload.TaskID
	uc.log.Infof("start processing export task: %s", taskID)

	// 1. 标记为 processing
	now := time.Now()
	if err := uc.repo.SaveTaskInfo(ctx, &exporttask.ExportTaskInfo{
		TaskID:    taskID,
		Status:    exporttask.ExportStatusProcessing,
		CreatedAt: now,
	}); err != nil {
		return uc.markFailed(ctx, taskID, fmt.Errorf("更新处理中状态失败: %w", err))
	}

	// 2. 分页拉取所有订单
	const pageSize int32 = 500
	var allOrders []*ExportOrder
	totalCount := int32(0)
	page := int32(1)

	for {
		orders, total, err := uc.orderRepo.ListOrdersForExport(ctx, page, pageSize, payload.Status, payload.DateFrom, payload.DateTo)
		if err != nil {
			return uc.markFailed(ctx, taskID, fmt.Errorf("查询订单失败: %w", err))
		}
		if totalCount == 0 {
			totalCount = total
			// 更新进度
			_ = uc.repo.UpdateTaskProgress(ctx, taskID, 0, totalCount)
		}
		allOrders = append(allOrders, orders...)
		// 更新已处理进度
		_ = uc.repo.UpdateTaskProgress(ctx, taskID, int32(len(allOrders)), totalCount)

		if int32(len(allOrders)) >= totalCount || len(orders) == 0 {
			break
		}
		page++
	}

	uc.log.Infof("export task %s: fetched %d orders", taskID, len(allOrders))

	// 3. 生成 Excel 并上传 OSS
	objectKey, err := uc.excelWriter.GenerateAndUpload(ctx, allOrders)
	if err != nil {
		return uc.markFailed(ctx, taskID, fmt.Errorf("生成并上传Excel失败: %w", err))
	}

	// 4. 标记为 completed
	finishedAt := time.Now()
	if err := uc.repo.SaveTaskInfo(ctx, &exporttask.ExportTaskInfo{
		TaskID:         taskID,
		Status:         exporttask.ExportStatusCompleted,
		ObjectKey:      objectKey,
		CreatedAt:      now,
		FinishedAt:      &finishedAt,
		TotalCount:     totalCount,
		ProcessedCount: int32(len(allOrders)),
	}); err != nil {
		return fmt.Errorf("更新完成状态失败: %w", err)
	}

	uc.log.Infof("export task completed: task_id=%s, object_key=%s", taskID, objectKey)
	return nil
}

// markFailed 标记任务失败
func (uc *ExportUseCase) markFailed(ctx context.Context, taskID string, cause error) error {
	uc.log.Errorf("export task failed: task_id=%s, err=%v", taskID, cause)
	finishedAt := time.Now()
	info, _ := uc.repo.GetTaskInfo(ctx, taskID)
	if info == nil {
		info = &exporttask.ExportTaskInfo{TaskID: taskID, CreatedAt: finishedAt}
	}
	info.Status = exporttask.ExportStatusFailed
	info.Error = cause.Error()
	info.FinishedAt = &finishedAt
	return uc.repo.SaveTaskInfo(ctx, info)
}
