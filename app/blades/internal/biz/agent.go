package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cyi-app/app/blades/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

const agentSessionPrefix = "session:"

// AgentUseCase 智能体用例（组合 AgentFactory + 会话管理）
type AgentUseCase struct {
	factory *AgentFactory
	rdb     redis.UniversalClient
	log     *log.Helper
}

// NewAgentUseCase 创建 AgentUseCase
func NewAgentUseCase(cfg *conf.Blades, rdb redis.UniversalClient, factory *AgentFactory, logger log.Logger) *AgentUseCase {
	return &AgentUseCase{
		factory: factory,
		rdb:     rdb,
		log:     log.NewHelper(logger),
	}
}

// sessionKey 会话 key
func sessionKey(source string, userID int64) string {
	return agentSessionPrefix + source + ":" + strconv.FormatInt(userID, 10)
}

type sessionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const sessionTTL = 2 * time.Hour

func (uc *AgentUseCase) loadHistory(ctx context.Context, source string, userID int64) ([]*sessionMessage, error) {
	if uc.rdb == nil {
		return nil, nil
	}
	data, err := uc.rdb.Get(ctx, sessionKey(source, userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	var msgs []*sessionMessage
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}
	// 只保留最近 10 轮
	if len(msgs) > 20 {
		msgs = msgs[len(msgs)-20:]
	}
	return msgs, nil
}

func (uc *AgentUseCase) saveHistory(ctx context.Context, source string, userID int64, userMsg, assistantMsg string) {
	if uc.rdb == nil {
		return
	}
	key := sessionKey(source, userID)
	var msgs []*sessionMessage
	data, err := uc.rdb.Get(ctx, key).Bytes()
	if err == nil {
		json.Unmarshal(data, &msgs)
	}
	msgs = append(msgs, &sessionMessage{Role: "user", Content: userMsg})
	msgs = append(msgs, &sessionMessage{Role: "assistant", Content: assistantMsg})
	if len(msgs) > 20 {
		msgs = msgs[len(msgs)-20:]
	}
	payload, _ := json.Marshal(msgs)
	uc.rdb.Set(ctx, key, payload, sessionTTL)
}

// Chat 非流式聊天
func (uc *AgentUseCase) Chat(ctx context.Context, userID int64, source, message string, enableRAG bool) (string, error) {
	inst, err := uc.factory.GetOrCreate(ctx, source)
	if err != nil {
		return "", fmt.Errorf("get agent failed: %w", err)
	}

	response, err := inst.Chat(ctx, message, enableRAG)
	if err != nil {
		return "", err
	}

	uc.saveHistory(ctx, source, userID, message, response)
	return response, nil
}

// ChatStream 流式聊天
func (uc *AgentUseCase) ChatStream(ctx context.Context, userID int64, source, message string, enableRAG bool, chunkCb func(string, bool)) error {
	inst, err := uc.factory.GetOrCreate(ctx, source)
	if err != nil {
		return fmt.Errorf("get agent failed: %w", err)
	}

	var fullResponse string
	err = inst.ChatStream(ctx, message, enableRAG, func(chunk string, done bool) {
		fullResponse += chunk
		chunkCb(chunk, done)
	})
	if err != nil {
		chunkCb("", true)
		return err
	}

	uc.saveHistory(ctx, source, userID, message, fullResponse)
	return nil
}

// ClearHistory 清除聊天历史
func (uc *AgentUseCase) ClearHistory(ctx context.Context, userID int64, source string) error {
	if uc.rdb == nil {
		return nil
	}
	return uc.rdb.Del(ctx, sessionKey(source, userID)).Err()
}

// HistoryMessage 历史消息
type HistoryMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// LoadChatHistory 加载聊天历史
func (uc *AgentUseCase) LoadChatHistory(ctx context.Context, userID int64, source string) ([]*HistoryMessage, error) {
	msgs, err := uc.loadHistory(ctx, source, userID)
	if err != nil {
		return nil, err
	}
	history := make([]*HistoryMessage, 0, len(msgs))
	for _, m := range msgs {
		history = append(history, &HistoryMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: time.Now().UnixMilli(),
		})
	}
	// 没有历史返回空数组
	if len(history) == 0 {
		return []*HistoryMessage{}, nil
	}
	return history, nil
}
