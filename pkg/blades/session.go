package blades

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kratos/blades"
	"github.com/redis/go-redis/v9"
)

const (
	sessionTTL    = 2 * time.Hour
	sessionPrefix = "chat:session:"
)

// sessionMessage Redis 中存储的聊天消息
type sessionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LoadHistory 从 Redis 加载会话历史
func LoadHistory(ctx context.Context, rdb redis.UniversalClient, sessionID string) ([]*blades.Message, error) {
	if sessionID == "" {
		return nil, nil
	}
	key := sessionPrefix + sessionID
	data, err := rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	var msgs []sessionMessage
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}
	history := make([]*blades.Message, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case "user":
			history = append(history, blades.UserMessage(m.Content))
		case "assistant":
			history = append(history, &blades.Message{
				Role:   blades.RoleAssistant,
				Parts:  blades.NewMessageParts(m.Content),
				Status: blades.StatusCompleted,
			})
		}
	}
	return history, nil
}

// SaveHistory 保存用户消息和 AI 回复到会话历史
func SaveHistory(ctx context.Context, rdb redis.UniversalClient, sessionID string, userMsg, assistantMsg string) error {
	if sessionID == "" {
		return nil
	}
	key := sessionPrefix + sessionID

	var msgs []sessionMessage
	data, err := rdb.Get(ctx, key).Bytes()
	if err == nil {
		json.Unmarshal(data, &msgs)
	}

	msgs = append(msgs, sessionMessage{Role: "user", Content: userMsg})
	msgs = append(msgs, sessionMessage{Role: "assistant", Content: assistantMsg})

	// 保留最近 20 轮对话
	if len(msgs) > 40 {
		msgs = msgs[len(msgs)-40:]
	}

	payload, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	return rdb.Set(ctx, key, payload, sessionTTL).Err()
}

// ClearHistory 清除会话历史
func ClearHistory(ctx context.Context, rdb redis.UniversalClient, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return rdb.Del(ctx, sessionPrefix+sessionID).Err()
}
