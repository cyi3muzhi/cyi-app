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
	sub  *nats.Subscription // 订阅句柄，用于优雅停机时取消订阅
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
	// NATS 连接失败（NewNATSConn 返回 nil）时降级跳过，不影响服务启动
	if c == nil || c.conn == nil {
		c.log.Warnf("NATS 连接不可用，跳过操作日志订阅")
		return
	}
	sub, err := c.conn.Subscribe(event.SubjectAdminOperation, func(msg *nats.Msg) {
		c.handle(msg.Data)
	})
	if err != nil {
		c.log.Errorf("订阅操作日志主题失败: %v", err)
		return
	}
	c.sub = sub
	c.log.Infof("已订阅操作日志主题: %s", event.SubjectAdminOperation)
}

// Stop 停止订阅（优雅停机时调用，取消订阅并等待在途消息处理完成）
func (c *OperationLogConsumer) Stop() {
	if c == nil || c.conn == nil || c.sub == nil {
		return
	}
	if err := c.sub.Unsubscribe(); err != nil {
		c.log.Warnf("取消操作日志订阅失败: %v", err)
		return
	}
	c.log.Infof("已取消操作日志主题订阅")
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
