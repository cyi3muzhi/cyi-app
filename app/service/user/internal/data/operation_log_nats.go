package data

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"cyi-app/app/service/user/internal/biz"
	"cyi-app/pkg/kratos/event"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
)

// OperationLogConsumer 操作日志 NATS 消费者
// 订阅 admin.operation 主题，将管理端操作日志事件写入数据库
type OperationLogConsumer struct {
	conn *nats.Conn
	uc   *biz.OperationLogUseCase
	log  *log.Helper
}

// NewOperationLogConsumer 创建操作日志消费者
func NewOperationLogConsumer(conn *nats.Conn, uc *biz.OperationLogUseCase, logger log.Logger) *OperationLogConsumer {
	return &OperationLogConsumer{
		conn: conn,
		uc:   uc,
		log:  log.NewHelper(logger),
	}
}

// Start 启动订阅（阻塞监听）
func (c *OperationLogConsumer) Start() {
	_, err := c.conn.Subscribe(event.SubjectAdminOperation, func(msg *nats.Msg) {
		c.handle(msg.Data)
	})
	if err != nil {
		c.log.Errorf("订阅操作日志主题失败: %v", err)
		return
	}
	c.log.Infof("已订阅操作日志主题: %s", event.SubjectAdminOperation)
}

// handle 处理一条操作日志事件
func (c *OperationLogConsumer) handle(data []byte) {
	ctx := context.Background()

	// panic 保护，防止单个消息处理崩溃导致订阅中断
	defer func() {
		if r := recover(); r != nil {
			c.log.Errorf("处理操作日志事件 panic: %v, stack: %s", r, string(debug.Stack()))
		}
	}()

	var evt event.AdminOperationLog
	if err := json.Unmarshal(data, &evt); err != nil {
		c.log.Errorf("解析操作日志事件失败: %v", err)
		return
	}

	record := &biz.OperationLogRecord{
		AdminUserID:  evt.AdminUserID,
		OperatorName: evt.OperatorName,
		RoleName:     evt.RoleName,
		Module:       evt.Module,
		Action:       evt.Action,
		Content:      evt.Content,
		Method:       evt.Method,
		Path:         evt.Path,
		IP:           evt.IP,
		UserAgent:    evt.UserAgent,
		RequestBody:  evt.RequestBody,
		Result:       evt.Result,
		StatusCode:   evt.StatusCode,
		DurationMs:   evt.DurationMs,
		CreatedAt:    evt.CreatedAt,
	}
	if err := c.uc.Create(ctx, record); err != nil {
		c.log.Errorf("写入操作日志失败: %v", err)
	}
}
