package data

import (
	"context"
	"encoding/json"
	"fmt"

	"cyi-app/app/task/internal/biz"
	"cyi-app/pkg/exporttask"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// exportRepo 导出任务状态仓库（基于 Redis 实现）
// task 消费端写入 processing/completed/failed 状态；admin 投递端写 pending。
// 两端通过共享包 exporttask 保持 key 前缀与结构一致。
type exportRepo struct {
	log   *log.Helper
	data  *Data
	cache redis.UniversalClient
}

// NewExportRepo 创建导出任务状态仓库
func NewExportRepo(data *Data, cache redis.UniversalClient, logger log.Logger) biz.ExportRepo {
	return &exportRepo{
		log:   log.NewHelper(logger),
		data:  data,
		cache: cache,
	}
}

// SaveTaskInfo 保存任务状态到 Redis（覆盖写，带 TTL）
func (r *exportRepo) SaveTaskInfo(ctx context.Context, info *exporttask.ExportTaskInfo) error {
	key := exporttask.TaskKey(info.TaskID)
	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return r.cache.Set(ctx, key, bytes, exporttask.TaskTTL).Err()
}

// GetTaskInfo 查询任务状态
func (r *exportRepo) GetTaskInfo(ctx context.Context, taskID string) (*exporttask.ExportTaskInfo, error) {
	key := exporttask.TaskKey(taskID)
	bytes, err := r.cache.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("任务不存在或已过期: %s", taskID)
	}
	if err != nil {
		return nil, err
	}
	var info exporttask.ExportTaskInfo
	if err := json.Unmarshal(bytes, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// UpdateTaskProgress 更新任务进度（读取后更新 processed_count 和 total_count，状态置为 processing）
func (r *exportRepo) UpdateTaskProgress(ctx context.Context, taskID string, processed, total int32) error {
	info, err := r.GetTaskInfo(ctx, taskID)
	if err != nil {
		return err
	}
	info.ProcessedCount = processed
	info.TotalCount = total
	info.Status = exporttask.ExportStatusProcessing
	return r.SaveTaskInfo(ctx, info)
}
