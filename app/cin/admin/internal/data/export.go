package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cyi-app/app/cin/admin/internal/biz"
	"cyi-app/pkg/exporttask"
	"cyi-app/pkg/util/rustfs"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// downloadURLExpiry 下载链接有效期（7 天，单位：秒）
// 与共享包 exporttask.DownloadURLExpiry 保持一致。
const downloadURLExpiry = exporttask.DownloadURLExpiry

// exportRepo 导出数据仓库实现（投递端）
// admin 投递任务时写 pending 状态并投递 asynq 任务；
// task 服务消费后写 processing/completed，admin 读取状态返回前端。
// 两端通过共享包 exporttask 保持 Redis key 与结构一致。
type exportRepo struct {
	log         *log.Helper
	data        *Data
	asynqClient *asynq.Client
}

// NewExportRepo 创建导出仓库
func NewExportRepo(data *Data, client *asynq.Client, logger log.Logger) biz.ExportRepo {
	return &exportRepo{
		log:         log.NewHelper(logger),
		data:        data,
		asynqClient: client,
	}
}

// EnqueueTask 投递导出任务到 Asynq "export" 队列（由 task 服务消费）
func (r *exportRepo) EnqueueTask(ctx context.Context, payload *exporttask.ExportTaskPayload) error {
	payloadBytes, err := exporttask.MarshalPayload(payload)
	if err != nil {
		return fmt.Errorf("序列化任务payload失败: %w", err)
	}
	// 注意：task type 必须用 ExportOrderTaskFullName（"async:export_order"），
	// 与 task 服务 AsyncHandleFunc 注册时经 AsyncName 转换后的完整名一致，
	// 否则会触发 "handler not found for task async:export_order"。
	task := asynq.NewTask(exporttask.ExportOrderTaskFullName, payloadBytes,
		asynq.Queue("export"),           // 导出队列（task 服务已配置消费）
		asynq.MaxRetry(2),               // 最多重试 2 次
		asynq.Timeout(10*time.Minute),   // 单任务最大执行时长 10 分钟
	)
	info, err := r.asynqClient.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("投递任务失败: %w", err)
	}
	r.log.Infof("export task enqueued: task_id=%s, asynq_id=%s", payload.TaskID, info.ID)
	return nil
}

// SaveTaskInfo 保存任务状态到 Redis（覆盖写，带 TTL）
func (r *exportRepo) SaveTaskInfo(ctx context.Context, info *exporttask.ExportTaskInfo) error {
	key := exporttask.TaskKey(info.TaskID)
	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return r.data.redis.Set(ctx, key, bytes, exporttask.TaskTTL).Err()
}

// GetTaskInfo 查询任务状态
func (r *exportRepo) GetTaskInfo(ctx context.Context, taskID string) (*exporttask.ExportTaskInfo, error) {
	key := exporttask.TaskKey(taskID)
	bytes, err := r.data.redis.Get(ctx, key).Bytes()
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

// GenerateDownloadURL 根据 objectKey 动态生成 presigned 下载 URL（7天有效期）
// admin 查询任务状态时，若任务 completed 则调用本方法生成临时下载链接返回前端。
func (r *exportRepo) GenerateDownloadURL(ctx context.Context, objectKey string) (string, error) {
	bucketName := string(rustfs.AdminServiceBucket)
	// 确保 bucket 存在
	exists, err := r.data.oss.BucketExists(ctx, bucketName)
	if err != nil {
		return "", fmt.Errorf("检查存储桶失败: %w", err)
	}
	if !exists {
		if err := r.data.oss.CreateBucket(ctx, bucketName); err != nil {
			return "", fmt.Errorf("创建存储桶失败: %w", err)
		}
	}
	return r.data.oss.GetPresignedURL(ctx, bucketName, objectKey, downloadURLExpiry)
}
