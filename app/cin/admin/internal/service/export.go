package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/pkg/exporttask"
)

// ExportOrder 导出订单（投递异步任务到 task 服务消费）
// 返回 task_id 供前端轮询任务状态
func (s *AdminService) ExportOrder(ctx context.Context, req *adminv1.ExportOrderRequest) (*adminv1.ExportOrderReply, error) {
	taskID, err := s.ec.SubmitExport(ctx, req.Status, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, err
	}
	return &adminv1.ExportOrderReply{
		TaskId: taskID,
		Status: adminv1.ExportStatus_PENDING,
	}, nil
}

// ExportTaskStatus 查询导出任务状态
// completed 时返回动态生成的下载链接（由 task 服务写回的 object_key 生成 presigned URL）
func (s *AdminService) ExportTaskStatus(ctx context.Context, req *adminv1.ExportTaskStatusRequest) (*adminv1.ExportTaskStatusReply, error) {
	info, downloadURL, err := s.ec.GetTaskStatus(ctx, req.TaskId)
	if err != nil {
		return nil, adminv1.ErrorNotFound("任务不存在或已过期")
	}

	reply := &adminv1.ExportTaskStatusReply{
		TaskId:         info.TaskID,
		Status:         mapBizStatusToProto(info.Status),
		DownloadUrl:    downloadURL,
		CreatedAt:      info.CreatedAt.Unix(),
		Error:          info.Error,
		TotalCount:     info.TotalCount,
		ProcessedCount: info.ProcessedCount,
	}
	if info.FinishedAt != nil {
		reply.FinishedAt = info.FinishedAt.Unix()
	}
	return reply, nil
}

// mapBizStatusToProto 将导出状态（共享包）映射到 proto 枚举
func mapBizStatusToProto(s exporttask.ExportStatus) adminv1.ExportStatus {
	switch s {
	case exporttask.ExportStatusPending:
		return adminv1.ExportStatus_PENDING
	case exporttask.ExportStatusProcessing:
		return adminv1.ExportStatus_PROCESSING
	case exporttask.ExportStatusCompleted:
		return adminv1.ExportStatus_COMPLETED
	case exporttask.ExportStatusFailed:
		return adminv1.ExportStatus_FAILED
	default:
		return adminv1.ExportStatus_ExportStatusUnspecified
	}
}
