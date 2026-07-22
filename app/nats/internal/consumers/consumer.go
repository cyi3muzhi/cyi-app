package consumers

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
)

// NatsConsumer NATS消息消费者
type NatsConsumer struct {
	conn     *nats.Conn
	handlers map[string]func(context.Context, []byte) error
	subs     []*nats.Subscription
	log      *log.Helper
}

// NewNatsConsumer 创建NATS消费者
func NewNatsConsumer(conn *nats.Conn, logger log.Logger) *NatsConsumer {
	return &NatsConsumer{
		conn:     conn,
		handlers: make(map[string]func(context.Context, []byte) error),
		log:      log.NewHelper(logger),
	}
}

// Subscribe 注册主题处理器
func (c *NatsConsumer) Subscribe(subject string, handler func(context.Context, []byte) error) {
	c.handlers[subject] = handler
}

// Start 启动所有订阅
func (c *NatsConsumer) Start() error {
	for subject, handler := range c.handlers {
		sub, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
			// 创建带 trace 的 context（后续接入 tracing 后自动继承 span）
			ctx := context.TODO()

			// panic 保护，防止单个 handler 崩溃导致整个订阅中断
			defer func() {
				if r := recover(); r != nil {
					c.log.WithContext(ctx).Errorf("NATS handler panic, subject: %s, recover: %v, stack: %s",
						msg.Subject, r, string(debug.Stack()))
				}
			}()

			c.log.WithContext(ctx).Infof("收到消息, subject: %s", msg.Subject)
			if err := handler(ctx, msg.Data); err != nil {
				c.log.WithContext(ctx).Errorf("处理消息失败, subject: %s, error: %v", msg.Subject, err)
			}
		})
		if err != nil {
			return fmt.Errorf("订阅失败, subject=%s: %w", subject, err)
		}
		c.subs = append(c.subs, sub)
		c.log.Infof("订阅主题: %s", subject)
	}
	return nil
}

// Stop 取消所有订阅
func (c *NatsConsumer) Stop() {
	for _, sub := range c.subs {
		if err := sub.Unsubscribe(); err != nil {
			c.log.Errorf("取消订阅失败: %v", err)
		}
	}
}
