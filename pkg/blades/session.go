package blades

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/blades"
	"github.com/redis/go-redis/v9"
)

const (
	defaultSessionTTL    = 2 * time.Hour
	defaultSessionPrefix = "chat:session:"
	defaultMaxRounds     = 20
)

// SessionMessage Redis 中存储的聊天消息
type SessionMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// SessionStore 会话存储，支持 Redis 和内存 fallback
type SessionStore struct {
	rdb       redis.UniversalClient
	localStor sync.Map // 内存会话存储（Redis 不可用时 fallback）
	prefix    string   // Redis key 前缀
	ttl       time.Duration
	maxRounds int // 最大对话轮数（每轮 = 1 条 user + 1 条 assistant）
}

// SessionOption 会话存储选项
type SessionOption func(*SessionStore)

// WithPrefix 设置 Redis key 前缀，默认 "chat:session:"
func WithPrefix(prefix string) SessionOption {
	return func(s *SessionStore) {
		s.prefix = prefix
	}
}

// WithTTL 设置会话 TTL，默认 2 小时
func WithTTL(ttl time.Duration) SessionOption {
	return func(s *SessionStore) {
		s.ttl = ttl
	}
}

// WithMaxRounds 设置最大对话轮数，默认 20 轮
func WithMaxRounds(max int) SessionOption {
	return func(s *SessionStore) {
		s.maxRounds = max
	}
}

// NewSessionStore 创建会话存储
// rdb 为 nil 或 Redis 不可用时自动降级为内存存储
func NewSessionStore(rdb redis.UniversalClient, opts ...SessionOption) *SessionStore {
	s := &SessionStore{
		rdb:       rdb,
		prefix:    defaultSessionPrefix,
		ttl:       defaultSessionTTL,
		maxRounds: defaultMaxRounds,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// key 生成完整的 Redis key
func (s *SessionStore) key(sessionID string) string {
	return s.prefix + sessionID
}

// LoadHistory 从会话加载历史，返回 blades.Message 切片
func (s *SessionStore) LoadHistory(ctx context.Context, sessionID string) ([]*blades.Message, error) {
	msgs, err := s.loadRaw(ctx, sessionID)
	if err != nil || msgs == nil {
		return nil, err
	}
	return s.toBladeMessages(msgs), nil
}

// LoadRawHistory 从会话加载原始消息（含 Timestamp）
func (s *SessionStore) LoadRawHistory(ctx context.Context, sessionID string) ([]SessionMessage, error) {
	return s.loadRaw(ctx, sessionID)
}

// loadRaw 从 Redis 或内存加载原始消息
func (s *SessionStore) loadRaw(ctx context.Context, sessionID string) ([]SessionMessage, error) {
	if sessionID == "" {
		return nil, nil
	}
	key := s.key(sessionID)

	// 先尝试从 Redis 加载
	if s.rdb != nil {
		data, err := s.rdb.Get(ctx, key).Bytes()
		if err == redis.Nil {
			return nil, nil
		}
		if err != nil {
			// Redis 出错时，尝试从内存加载
			return s.loadFromMemory(key), nil
		}
		var msgs []SessionMessage
		if err := json.Unmarshal(data, &msgs); err != nil {
			return nil, fmt.Errorf("unmarshal failed: %w", err)
		}
		if len(msgs) > s.maxRounds*2 {
			msgs = msgs[len(msgs)-s.maxRounds*2:]
		}
		return msgs, nil
	}

	// 从内存加载
	return s.loadFromMemory(key), nil
}

// loadFromMemory 从内存加载消息
func (s *SessionStore) loadFromMemory(key string) []SessionMessage {
	if val, ok := s.localStor.Load(key); ok {
		msgs := val.([]SessionMessage)
		if len(msgs) > s.maxRounds*2 {
			msgs = msgs[len(msgs)-s.maxRounds*2:]
		}
		return msgs
	}
	return nil
}

// toBladeMessages 将 SessionMessage 转为 blades.Message 切片
func (s *SessionStore) toBladeMessages(msgs []SessionMessage) []*blades.Message {
	var result []*blades.Message
	for _, m := range msgs {
		switch m.Role {
		case "user":
			result = append(result, blades.UserMessage(m.Content))
		case "assistant":
			result = append(result, &blades.Message{
				Role:   blades.RoleAssistant,
				Parts:  blades.NewMessageParts(m.Content),
				Status: blades.StatusCompleted,
			})
		}
	}
	return result
}

// SaveHistory 保存用户消息和 AI 回复到会话历史
func (s *SessionStore) SaveHistory(ctx context.Context, sessionID string, userMsg, assistantMsg string) error {
	if sessionID == "" {
		return nil
	}
	key := s.key(sessionID)

	// 读取已有的历史
	var msgs []SessionMessage
	if s.rdb != nil {
		data, err := s.rdb.Get(context.Background(), key).Bytes()
		if err == nil {
			json.Unmarshal(data, &msgs)
		}
	} else if val, ok := s.localStor.Load(key); ok {
		msgs = val.([]SessionMessage)
	}

	// 追加新消息
	now := time.Now().UnixMilli()
	msgs = append(msgs, SessionMessage{Role: "user", Content: userMsg, Timestamp: now})
	msgs = append(msgs, SessionMessage{Role: "assistant", Content: assistantMsg, Timestamp: now})

	// 保留最近 N 轮对话
	if len(msgs) > s.maxRounds*2 {
		msgs = msgs[len(msgs)-s.maxRounds*2:]
	}

	payload, _ := json.Marshal(msgs)

	// 存到 Redis（优先）或内存
	if s.rdb != nil {
		return s.rdb.Set(context.Background(), key, payload, s.ttl).Err()
	}
	s.localStor.Store(key, msgs)
	return nil
}

// ClearHistory 清除会话历史
func (s *SessionStore) ClearHistory(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	key := s.key(sessionID)
	if s.rdb != nil {
		return s.rdb.Del(ctx, key).Err()
	}
	s.localStor.Delete(key)
	return nil
}

// --- 以下为向后兼容的包级函数 ---

// LoadHistory 从 Redis 加载会话历史（向后兼容的包级函数）
// 仅支持 Redis 存储，无内存 fallback
func LoadHistory(ctx context.Context, rdb redis.UniversalClient, sessionID string) ([]*blades.Message, error) {
	if sessionID == "" {
		return nil, nil
	}
	key := defaultSessionPrefix + sessionID
	data, err := rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	var msgs []SessionMessage
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

// SaveHistory 保存用户消息和 AI 回复到会话历史（向后兼容的包级函数）
func SaveHistory(ctx context.Context, rdb redis.UniversalClient, sessionID string, userMsg, assistantMsg string) error {
	if sessionID == "" {
		return nil
	}
	key := defaultSessionPrefix + sessionID

	var msgs []SessionMessage
	data, err := rdb.Get(ctx, key).Bytes()
	if err == nil {
		json.Unmarshal(data, &msgs)
	}

	msgs = append(msgs, SessionMessage{Role: "user", Content: userMsg})
	msgs = append(msgs, SessionMessage{Role: "assistant", Content: assistantMsg})

	// 保留最近 20 轮对话
	if len(msgs) > defaultMaxRounds*2 {
		msgs = msgs[len(msgs)-defaultMaxRounds*2:]
	}

	payload, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	return rdb.Set(ctx, key, payload, defaultSessionTTL).Err()
}

// ClearHistory 清除会话历史（向后兼容的包级函数）
func ClearHistory(ctx context.Context, rdb redis.UniversalClient, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return rdb.Del(ctx, defaultSessionPrefix+sessionID).Err()
}
