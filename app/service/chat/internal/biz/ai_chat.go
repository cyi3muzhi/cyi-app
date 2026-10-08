package biz

import (
	"context"
	"time"

	bladesRpc "cyi-app/api/blades/v1"
	bladesclient "cyi-app/pkg/blades/client"
	"cyi-app/pkg/kratos/event"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// aiSource 客服 AI 兜底在 blades 服务中的来源标识（与用户端 AI 聊天共用 user 来源的历史隔离维度）
const aiSource = "user"

// AIChatUseCase AI 客服兜底用例
// 会话处于「待接入 + AI 接待中」时，用户发消息由本用例异步调用 blades 生成回复，
// 以 AI 消息（sender_type=4）落库并推送，客服接入后会话切换人工接待即停止。
type AIChatUseCase struct {
	ai       *bladesclient.Client
	convRepo ConversationRepo
	msgRepo  MessageRepo
	pub      ChatMessagePublisher
	rdb      redis.UniversalClient
	log      *log.Helper
}

// NewAIChatUseCase 创建 AI 客服兜底用例
func NewAIChatUseCase(bc bladesRpc.BladesClient, convRepo ConversationRepo, msgRepo MessageRepo, pub ChatMessagePublisher, rdb redis.UniversalClient, logger log.Logger) *AIChatUseCase {
	return &AIChatUseCase{
		ai:       bladesclient.NewClient(bc, logger),
		convRepo: convRepo,
		msgRepo:  msgRepo,
		pub:      pub,
		rdb:      rdb,
		log:      log.NewHelper(logger),
	}
}

// HandleAIReply 用户发送消息后异步触发 AI 回复（仅 AI 接待中会话生效）
func (uc *AIChatUseCase) HandleAIReply(ctx context.Context, conversationID, userID int64, userMessage string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		// 重新读取会话，确认仍处于「待接入 + AI 接待中」（客服可能已接入）
		conv, err := uc.convRepo.GetConversation(ctx, conversationID)
		if err != nil || conv == nil {
			return
		}
		if conv.Status != ConversationStatusPending || conv.ServiceType != ServiceTypeAI {
			return
		}

		// 调用 blades 生成 AI 回复（非流式，复用该用户历史上下文）
		reply, err := uc.ai.Chat(ctx, aiSource, userID, "", userMessage, false)
		if err != nil || reply == "" {
			uc.log.Warnf("ai fallback chat failed: %v", err)
			return
		}

		// 落库 AI 消息
		now := time.Now().Unix()
		msg := &ChatMessage{
			ConversationID: conversationID,
			SenderType:     SenderTypeAI,
			SenderID:       0,
			MsgType:        MsgTypeText,
			Content:        reply,
			IsRead:         0,
		}
		if err := uc.msgRepo.CreateMessage(ctx, msg); err != nil {
			uc.log.Warnf("ai fallback save message failed: %v", err)
			return
		}

		// 更新会话摘要
		summary := reply
		if len([]rune(summary)) > 100 {
			summary = string([]rune(summary)[:100]) + "..."
		}
		if err := uc.convRepo.UpdateLastMessage(ctx, conversationID, MsgTypeText, summary, now); err != nil {
			uc.log.Warnf("ai fallback update last message failed: %v", err)
		}

		// 发布推送事件（AI 消息推送给用户，并对端未读 +1）
		IncrUnread(ctx, uc.rdb, conversationID, event.ChatReceiverUser)
		if uc.pub != nil {
			uc.pub.PublishMessage(ctx, &event.ChatMessageEvent{
				ConversationID: conversationID,
				MessageID:      msg.ID,
				SenderType:     SenderTypeAI,
				SenderID:       0,
				ReceiverType:   event.ChatReceiverUser,
				ReceiverID:     userID,
				MsgType:        MsgTypeText,
				Content:        reply,
				CreatedAt:      now,
			})
		}
	}()
}
