// Package exporttask 提供订单导出异步任务的共享类型与常量。
// admin 服务（投递端）与 task 服务（消费端）共同引用本包，
// 避免跨服务反向依赖 biz 包，并保证两端读写 Redis 任务状态时结构一致。
package exporttask

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExportOrderTaskName 导出任务的逻辑名称（不含 async: 前缀）
// task 服务通过 AsyncHandleFunc 注册时使用本常量，transporter.AsyncName 会自动补 "async:" 前缀。
// 注意：不要在此带 "async:" 前缀，否则会变成 "async:async:export_order" 导致 handler not found。
const ExportOrderTaskName = "export_order"

// ExportOrderTaskFullName 导出任务在 asynq 队列中的完整类型名
// admin 投递任务时使用（asynq.NewTask 的第一个参数）。
// 格式 "async:{name}" 与 transporter.AsyncName(ExportOrderTaskName) 的产物保持一致，
// 保证 task 注册的 handler name 与 admin 投递的 task type 完全对齐。
const ExportOrderTaskFullName = "async:" + ExportOrderTaskName

// Redis key 前缀与保留时长
const (
	// RedisKeyPrefix 任务状态在 Redis 中的 key 前缀，最终 key = RedisKeyPrefix + taskID
	// admin 写 pending，task 写 processing/completed，两端必须一致。
	RedisKeyPrefix = "admin:export:"
	// TaskTTL 任务状态在 Redis 中的保留时长（24 小时）
	TaskTTL = 24 * time.Hour
	// DownloadURLExpiry 下载链接有效期（7 天，单位：秒）
	DownloadURLExpiry = 7 * 24 * 3600
)

// ExportStatus 导出任务状态
type ExportStatus string

const (
	ExportStatusPending    ExportStatus = "pending"
	ExportStatusProcessing ExportStatus = "processing"
	ExportStatusCompleted ExportStatus = "completed"
	ExportStatusFailed     ExportStatus = "failed"
)

// ExportTaskPayload Asynq 任务 payload（序列化为 JSON 传递给 worker）
// admin 投递时构造，task 消费时反序列化。
type ExportTaskPayload struct {
	TaskID   string `json:"task_id"`
	Status   string `json:"status"`    // 订单状态筛选条件（管理端编码，如 paid/shipped）
	DateFrom string `json:"date_from"` // 起始日期 "2006-01-02"
	DateTo   string `json:"date_to"`   // 结束日期 "2006-01-02"
}

// ExportTaskInfo 任务状态信息（存储在 Redis）
// admin 写初始 pending，task 写 processing/completed/failed，admin 读取返回前端。
type ExportTaskInfo struct {
	TaskID         string       `json:"task_id"`
	Status         ExportStatus `json:"status"`
	ObjectKey      string       `json:"object_key,omitempty"` // OSS 对象路径（不存 URL 避免过期）
	CreatedAt      time.Time    `json:"created_at"`
	FinishedAt     *time.Time   `json:"finished_at,omitempty"`
	Error          string       `json:"error,omitempty"`
	TotalCount     int32        `json:"total_count,omitempty"`
	ProcessedCount int32        `json:"processed_count,omitempty"`
}

// TaskKey 拼接任务在 Redis 中的完整 key
func TaskKey(taskID string) string {
	return RedisKeyPrefix + taskID
}

// MarshalPayload 序列化任务 payload（供 admin 投递 asynq.NewTask 时使用）
func MarshalPayload(p *ExportTaskPayload) ([]byte, error) {
	return json.Marshal(p)
}

// UnmarshalPayload 反序列化任务 payload（供 task 消费 asynq.Task.Payload() 时使用）
func UnmarshalPayload(b []byte) (*ExportTaskPayload, error) {
	var p ExportTaskPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("解析export payload失败: %w", err)
	}
	return &p, nil
}
