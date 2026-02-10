package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

// MessageHandler 消息处理器接口
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg []byte) error
}

// MessageRouter 消息路由器
type MessageRouter struct {
	handlers map[string]MessageHandler
	log      *log.Helper
}

// NewMessageRouter 创建消息路由器
func NewMessageRouter(logger log.Logger) *MessageRouter {
	return &MessageRouter{
		handlers: make(map[string]MessageHandler),
		log:      log.NewHelper(logger),
	}
}

// RegisterHandler 注册消息处理器
func (r *MessageRouter) RegisterHandler(messageType string, handler MessageHandler) {
	r.handlers[messageType] = handler
}

// HandleMessage 处理消息
func (r *MessageRouter) HandleMessage(ctx context.Context, messageType string, msg []byte) error {
	handler, ok := r.handlers[messageType]
	if !ok {
		r.log.Error("No handler registered for message type: %s", messageType)
		return fmt.Errorf("no handler registered for message type: %s", messageType)
	}

	return handler.HandleMessage(ctx, msg)
}

// TaskMessage 任务消息结构
type TaskMessage struct {
	Action string          `json:"action"` // create, update, delete
	Data   json.RawMessage `json:"data"`   // 任务数据
}

// TaskService 任务服务接口
type TaskService interface {
	CreateTask(ctx context.Context, title, description, status string) error
	UpdateTask(ctx context.Context, id int64, title, description, status string) error
	DeleteTask(ctx context.Context, id int64) error
}

// TaskMessageHandler 任务消息处理器
type TaskMessageHandler struct {
	taskService TaskService
	log         *log.Helper
}

// NewTaskMessageHandler 创建任务消息处理器
func NewTaskMessageHandler(taskService interface{}, logger log.Logger) *TaskMessageHandler {
	// 尝试将taskService转换为TaskService接口
	ts, ok := taskService.(TaskService)
	if !ok {
		log.NewHelper(logger).Warn("taskService does not implement TaskService interface, task operations will be skipped")
	}

	return &TaskMessageHandler{
		taskService: ts,
		log:         log.NewHelper(logger),
	}
}

// HandleMessage 处理任务消息
func (h *TaskMessageHandler) HandleMessage(ctx context.Context, msg []byte) error {
	var taskMsg TaskMessage
	if err := json.Unmarshal(msg, &taskMsg); err != nil {
		return fmt.Errorf("failed to unmarshal task message: %w", err)
	}

	switch taskMsg.Action {
	case "create":
		return h.handleCreateTask(ctx, taskMsg.Data)
	case "update":
		return h.handleUpdateTask(ctx, taskMsg.Data)
	case "delete":
		return h.handleDeleteTask(ctx, taskMsg.Data)
	default:
		return fmt.Errorf("unknown task action: %s", taskMsg.Action)
	}
}

// handleCreateTask 处理创建任务消息
func (h *TaskMessageHandler) handleCreateTask(ctx context.Context, data json.RawMessage) error {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to unmarshal create task request: %w", err)
	}

	// 调用服务创建任务
	if h.taskService != nil {
		if err := h.taskService.CreateTask(ctx, req.Title, req.Description, req.Status); err != nil {
			h.log.Error("Failed to create task: %v", err)
			return err
		}
	}

	h.log.Info("Created task: %s", req.Title)
	return nil
}

// handleUpdateTask 处理更新任务消息
func (h *TaskMessageHandler) handleUpdateTask(ctx context.Context, data json.RawMessage) error {
	var req struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to unmarshal update task request: %w", err)
	}

	// 调用服务更新任务
	if h.taskService != nil {
		if err := h.taskService.UpdateTask(ctx, req.ID, req.Title, req.Description, req.Status); err != nil {
			h.log.Error("Failed to update task: %v", err)
			return err
		}
	}

	h.log.Info("Updated task: %d", req.ID)
	return nil
}

// handleDeleteTask 处理删除任务消息
func (h *TaskMessageHandler) handleDeleteTask(ctx context.Context, data json.RawMessage) error {
	var req struct {
		ID int64 `json:"id"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to unmarshal delete task request: %w", err)
	}

	// 调用服务删除任务
	if h.taskService != nil {
		if err := h.taskService.DeleteTask(ctx, req.ID); err != nil {
			h.log.Error("Failed to delete task: %v", err)
			return err
		}
	}

	h.log.Info("Deleted task: %d", req.ID)
	return nil
}
