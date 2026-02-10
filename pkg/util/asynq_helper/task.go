package asynqhelper

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

// TaskType 定义任务类型
const (
	TaskTypeEmail       = "email:send"
	TaskTypeSMS         = "sms:send"
	TaskTypeNotification = "notification:send"
	TaskTypeDataSync    = "data:sync"
	TaskTypeImageProcess = "image:process"
)

// EmailTaskPayload 邮件任务的负载
type EmailTaskPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// SMSTaskPayload 短信任务的负载
type SMSTaskPayload struct {
	Phone   string `json:"phone"`
	Content string `json:"content"`
}

// NotificationTaskPayload 通知任务的负载
type NotificationTaskPayload struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"` // 通知类型: info, warning, error, success
}

// DataSyncTaskPayload 数据同步任务的负载
type DataSyncTaskPayload struct {
	Source string `json:"source"` // 数据源
	Target string `json:"target"` // 数据目标
	Filter string `json:"filter"` // 过滤条件
}

// ImageProcessTaskPayload 图片处理任务的负载
type ImageProcessTaskPayload struct {
	ImageID   string `json:"image_id"`
	Operation string `json:"operation"` // 操作类型: resize, crop, compress
	Params    map[string]interface{} `json:"params"` // 操作参数
	Output    string `json:"output"` // 输出路径
}

// CreateEmailTask 创建邮件任务
func (c *Client) CreateEmailTask(payload *EmailTaskPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建任务
	task := asynq.NewTask(TaskTypeEmail, data, opts...)

	// 提交任务
	info, err := c.client.Enqueue(task)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// CreateSMSTask 创建短信任务
func (c *Client) CreateSMSTask(payload *SMSTaskPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建任务
	task := asynq.NewTask(TaskTypeSMS, data, opts...)

	// 提交任务
	info, err := c.client.Enqueue(task)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// CreateNotificationTask 创建通知任务
func (c *Client) CreateNotificationTask(payload *NotificationTaskPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建任务
	task := asynq.NewTask(TaskTypeNotification, data, opts...)

	// 提交任务
	info, err := c.client.Enqueue(task)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// CreateDataSyncTask 创建数据同步任务
func (c *Client) CreateDataSyncTask(payload *DataSyncTaskPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建任务
	task := asynq.NewTask(TaskTypeDataSync, data, opts...)

	// 提交任务
	info, err := c.client.Enqueue(task)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// CreateImageProcessTask 创建图片处理任务
func (c *Client) CreateImageProcessTask(payload *ImageProcessTaskPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建任务
	task := asynq.NewTask(TaskTypeImageProcess, data, opts...)

	// 提交任务
	info, err := c.client.Enqueue(task)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// CreateTask 创建自定义类型的任务
func (c *Client) CreateTask(taskType string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建任务
	task := asynq.NewTask(taskType, data, opts...)

	// 提交任务
	info, err := c.client.Enqueue(task)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// ScheduleTask 调度延迟任务
func (c *Client) ScheduleTask(taskType string, payload interface{}, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建任务
	task := asynq.NewTask(taskType, data)

	// 提交任务
	info, err := c.client.Enqueue(task, append(opts, asynq.ProcessIn(delay))...)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// ScheduleTaskAt 调度在指定时间执行的任务
func (c *Client) ScheduleTaskAt(taskType string, payload interface{}, at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建任务
	task := asynq.NewTask(taskType, data)

	// 提交任务
	info, err := c.client.Enqueue(task, append(opts, asynq.ProcessAt(at))...)
	if err != nil {
		return nil, err
	}

	return info, nil
}
