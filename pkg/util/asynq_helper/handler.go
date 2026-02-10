package asynqhelper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// Handler 任务处理器接口
type Handler interface {
	// ProcessTask 处理任务
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// TaskHandlerFunc 任务处理函数类型
type TaskHandlerFunc func(ctx context.Context, task *asynq.Task) error

// ProcessTask 实现Handler接口
func (f TaskHandlerFunc) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return f(ctx, task)
}

// MultiHandler 多任务处理器，根据任务类型分发任务
type MultiHandler struct {
	handlers map[string]Handler
}

// NewMultiHandler 创建一个新的多任务处理器
func NewMultiHandler() *MultiHandler {
	return &MultiHandler{
		handlers: make(map[string]Handler),
	}
}

// RegisterHandler 注册任务处理器
func (h *MultiHandler) RegisterHandler(taskType string, handler Handler) {
	h.handlers[taskType] = handler
}

// RegisterHandlerFunc 注册任务处理函数
func (h *MultiHandler) RegisterHandlerFunc(taskType string, fn TaskHandlerFunc) {
	h.handlers[taskType] = fn
}

// ProcessTask 处理任务，根据任务类型分发到对应的处理器
func (h *MultiHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	handler, ok := h.handlers[task.Type()]
	if !ok {
		return fmt.Errorf("no handler registered for task type: %s", task.Type())
	}

	return handler.ProcessTask(ctx, task)
}

// DefaultHandler 默认任务处理器，提供一些常见任务的处理实现
type DefaultHandler struct {
	handler *MultiHandler
}

// NewDefaultHandler 创建一个新的默认任务处理器
func NewDefaultHandler() *DefaultHandler {
	h := &DefaultHandler{
		handler: NewMultiHandler(),
	}

	// 注册默认的任务处理器
	h.RegisterDefaultHandlers()

	return h
}

// RegisterDefaultHandlers 注册默认的任务处理器
func (h *DefaultHandler) RegisterDefaultHandlers() {
	// 注册邮件任务处理器
	h.handler.RegisterHandlerFunc(TaskTypeEmail, func(ctx context.Context, task *asynq.Task) error {
		var payload EmailTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal email task payload: %w", err)
		}

		// 这里是邮件发送的实现
		// 实际应用中，你需要替换为真正的邮件发送逻辑
		fmt.Printf("Sending email to %s: %s\n", payload.To, payload.Subject)

		return nil
	})

	// 注册短信任务处理器
	h.handler.RegisterHandlerFunc(TaskTypeSMS, func(ctx context.Context, task *asynq.Task) error {
		var payload SMSTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal sms task payload: %w", err)
		}

		// 这里是短信发送的实现
		// 实际应用中，你需要替换为真正的短信发送逻辑
		fmt.Printf("Sending SMS to %s: %s\n", payload.Phone, payload.Content)

		return nil
	})

	// 注册通知任务处理器
	h.handler.RegisterHandlerFunc(TaskTypeNotification, func(ctx context.Context, task *asynq.Task) error {
		var payload NotificationTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal notification task payload: %w", err)
		}

		// 这里是通知发送的实现
		// 实际应用中，你需要替换为真正的通知发送逻辑
		fmt.Printf("Sending notification to user %s: %s\n", payload.UserID, payload.Title)

		return nil
	})

	// 注册数据同步任务处理器
	h.handler.RegisterHandlerFunc(TaskTypeDataSync, func(ctx context.Context, task *asynq.Task) error {
		var payload DataSyncTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal data sync task payload: %w", err)
		}

		// 这里是数据同步的实现
		// 实际应用中，你需要替换为真正的数据同步逻辑
		fmt.Printf("Syncing data from %s to %s\n", payload.Source, payload.Target)

		return nil
	})

	// 注册图片处理任务处理器
	h.handler.RegisterHandlerFunc(TaskTypeImageProcess, func(ctx context.Context, task *asynq.Task) error {
		var payload ImageProcessTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal image process task payload: %w", err)
		}

		// 这里是图片处理的实现
		// 实际应用中，你需要替换为真正的图片处理逻辑
		fmt.Printf("Processing image %s with operation %s\n", payload.ImageID, payload.Operation)

		return nil
	})
}

// RegisterHandler 注册自定义任务处理器
func (h *DefaultHandler) RegisterHandler(taskType string, handler Handler) {
	h.handler.RegisterHandler(taskType, handler)
}

// RegisterHandlerFunc 注册自定义任务处理函数
func (h *DefaultHandler) RegisterHandlerFunc(taskType string, fn TaskHandlerFunc) {
	h.handler.RegisterHandlerFunc(taskType, fn)
}

// ProcessTask 处理任务
func (h *DefaultHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return h.handler.ProcessTask(ctx, task)
}
