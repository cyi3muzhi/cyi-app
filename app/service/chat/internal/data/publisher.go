package data

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"cyi-app/app/service/chat/internal/biz"
	"cyi-app/pkg/kratos/event"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
)

// chatMessagePublisher 聊天消息 NATS 发布器
// 消息落库后由 biz 调用，将事件发布到 chat.message 主题，网关订阅后实时推送。
type chatMessagePublisher struct {
	conn *nats.Conn
	log  *log.Helper
}

// NewChatMessagePublisher 创建聊天消息发布器
func NewChatMessagePublisher(conn *nats.Conn, logger log.Logger) biz.ChatMessagePublisher {
	return &chatMessagePublisher{conn: conn, log: log.NewHelper(logger)}
}

// PublishMessage 发布一条消息事件（异步，失败仅记录日志不影响主流程）
func (p *chatMessagePublisher) PublishMessage(ctx context.Context, evt *event.ChatMessageEvent) {
	if p == nil || p.conn == nil || evt == nil {
		return
	}
	data, err := json.Marshal(evt)
	if err != nil {
		p.log.Errorf("序列化聊天消息事件失败: %v", err)
		return
	}
	if err := p.conn.Publish(event.SubjectChatMessage, data); err != nil {
		p.log.Errorf("发布聊天消息事件失败: %v", err)
	}
}

// conversationEventPublisher 会话状态事件发布器
// 会话新建/接入/转接/结束时发布到 chat.conversation 主题，供客服工作台实时刷新与用户端状态通知。
type conversationEventPublisher struct {
	conn *nats.Conn
	log  *log.Helper
}

// NewConversationEventPublisher 创建会话状态事件发布器
func NewConversationEventPublisher(conn *nats.Conn, logger log.Logger) biz.ConversationEventPublisher {
	return &conversationEventPublisher{conn: conn, log: log.NewHelper(logger)}
}

// PublishConversation 发布一条会话状态事件（异步，失败仅记录日志不影响主流程）
func (p *conversationEventPublisher) PublishConversation(ctx context.Context, evt *event.ChatConversationEvent) {
	if p == nil || p.conn == nil || evt == nil {
		return
	}
	data, err := json.Marshal(evt)
	if err != nil {
		p.log.Errorf("序列化会话状态事件失败: %v", err)
		return
	}
	if err := p.conn.Publish(event.SubjectChatConversation, data); err != nil {
		p.log.Errorf("发布会话状态事件失败: %v", err)
	}
}

// NewNATSConn 创建 NATS 连接（单例，最多重试 5 次，失败降级为 nil 不 panic）
// 优先从环境变量 NATS_URL 读取地址（Docker 环境配置为 nats://nats:4222），默认 localhost。
func NewNATSConn() *nats.Conn {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	for i := 0; i < 5; i++ {
		nc, err := nats.Connect(natsURL)
		if err == nil {
			return nc
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}
