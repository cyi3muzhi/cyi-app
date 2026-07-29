package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"cyi-app/app/blades/internal/conf"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

const (
	agentSessionPrefix = "session:"
	preferencePrefix   = "preference:"
)

// contextKey 用于 context.WithValue 的 key 类型
type contextKey string

const (
	userIDContextKey contextKey = "user_id"
	sourceContextKey contextKey = "source"
	roleContextKey   contextKey = "role"
)

// AgentUseCase 智能体用例（组合 AgentFactory + 会话管理）
type AgentUseCase struct {
	factory   *AgentFactory
	rdb       redis.UniversalClient
	localStor sync.Map // 内存会话存储（Redis 不可用时 fallback）
	log       *log.Helper
}

// NewAgentUseCase 创建 AgentUseCase
func NewAgentUseCase(cfg *conf.Blades, rdb redis.UniversalClient, factory *AgentFactory, logger log.Logger) *AgentUseCase {
	rdbAvailable := rdb != nil
	if rdbAvailable {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			rdbAvailable = false
			log.NewHelper(logger).Warnf("redis unavailable, falling back to local memory storage: %v", err)
		}
	}
	if !rdbAvailable {
		rdb = nil
	}
	return &AgentUseCase{
		factory: factory,
		rdb:     rdb,
		log:     log.NewHelper(logger),
	}
}

// preferenceKey 偏好存储 key
func preferenceKey(source string, userID int64) string {
	return preferencePrefix + source + ":" + strconv.FormatInt(userID, 10)
}

// loadPreferences 加载用户偏好，返回格式化文本
func (uc *AgentUseCase) loadPreferences(ctx context.Context, source string, userID int64) string {
	key := preferenceKey(source, userID)

	var prefs map[string]string
	if uc.rdb != nil {
		data, err := uc.rdb.Get(ctx, key).Bytes()
		if err == nil {
			json.Unmarshal(data, &prefs)
		}
	} else if val, ok := uc.localStor.Load(key); ok {
		prefs = val.(map[string]string)
	}

	if len(prefs) == 0 {
		return ""
	}

	var text string
	for k, v := range prefs {
		text += fmt.Sprintf("- %s: %s\n", k, v)
	}
	return text
}

// savePreference 保存单条用户偏好
func (uc *AgentUseCase) savePreference(ctx context.Context, source string, userID int64, key_, value string) error {
	key := preferenceKey(source, userID)

	var prefs map[string]string
	if uc.rdb != nil {
		data, err := uc.rdb.Get(ctx, key).Bytes()
		if err == nil {
			json.Unmarshal(data, &prefs)
		}
	} else if val, ok := uc.localStor.Load(key); ok {
		prefs = val.(map[string]string)
	}

	if prefs == nil {
		prefs = make(map[string]string)
	}
	prefs[key_] = value

	payload, _ := json.Marshal(prefs)
	if uc.rdb != nil {
		return uc.rdb.Set(ctx, key, payload, 0).Err()
	}
	uc.localStor.Store(key, prefs)
	return nil
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
	key := sessionKey(source, userID)

	// 先尝试从 Redis 加载
	if uc.rdb != nil {
		data, err := uc.rdb.Get(ctx, key).Bytes()
		if err == redis.Nil {
			return nil, nil
		}
		if err != nil {
			uc.log.Warnf("redis load history failed for key=%s: %v, fallback to local", key, err)
		} else {
			var msgs []*sessionMessage
			if err := json.Unmarshal(data, &msgs); err != nil {
				return nil, fmt.Errorf("unmarshal failed: %w", err)
			}
			if len(msgs) > 20 {
				msgs = msgs[len(msgs)-20:]
			}
			return msgs, nil
		}
	}

	// fallback: 从内存加载
	if val, ok := uc.localStor.Load(key); ok {
		msgs := val.([]*sessionMessage)
		if len(msgs) > 20 {
			msgs = msgs[len(msgs)-20:]
		}
		return msgs, nil
	}
	return nil, nil
}

// sessionToBladeMessages 将 session 消息转为 blades.Message 切片
func (uc *AgentUseCase) sessionToBladeMessages(msgs []*sessionMessage) []*blades.Message {
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

func (uc *AgentUseCase) saveHistory(ctx context.Context, source string, userID int64, userMsg, assistantMsg string) {
	key := sessionKey(source, userID)

	// 读取已有的历史
	var msgs []*sessionMessage
	if uc.rdb != nil {
		// 使用 Background context 避免请求结束时 context canceled
		data, err := uc.rdb.Get(context.Background(), key).Bytes()
		if err == nil {
			json.Unmarshal(data, &msgs)
		}
	} else if val, ok := uc.localStor.Load(key); ok {
		msgs = val.([]*sessionMessage)
	}

	// 追加新消息
	msgs = append(msgs, &sessionMessage{Role: "user", Content: userMsg})
	msgs = append(msgs, &sessionMessage{Role: "assistant", Content: assistantMsg})
	if len(msgs) > 20 {
		msgs = msgs[len(msgs)-20:]
	}

	payload, _ := json.Marshal(msgs)

	// 存到 Redis（优先）或内存
	if uc.rdb != nil {
		uc.rdb.Set(context.Background(), key, payload, sessionTTL)
	} else {
		uc.localStor.Store(key, msgs)
	}
}

// Chat 非流式聊天
func (uc *AgentUseCase) Chat(ctx context.Context, userID int64, source, role, message string, enableRAG bool) (string, error) {
	inst, err := uc.factory.GetOrCreate(ctx, source, role)
	if err != nil {
		return "", fmt.Errorf("get agent failed: %w", err)
	}

	// 加载会话历史并传递给 Agent
	history, _ := uc.loadHistory(ctx, source, userID)
	historyMessages := uc.sessionToBladeMessages(history)
	var instruction *blades.Message
	if prefText := uc.loadPreferences(ctx, source, userID); prefText != "" {
		instruction = blades.SystemMessage("以下是该用户的偏好信息，请遵守：\n" + prefText)
	}

	// 将 userID/source/role 放入 context，供工具函数使用
	ctx = context.WithValue(ctx, userIDContextKey, userID)
	ctx = context.WithValue(ctx, sourceContextKey, source)
	ctx = context.WithValue(ctx, roleContextKey, role)

	response, err := inst.Chat(ctx, message, enableRAG, historyMessages, instruction)
	if err != nil {
		return "", err
	}

	uc.saveHistory(ctx, source, userID, message, response)
	return response, nil
}

// ChatStream 流式聊天
func (uc *AgentUseCase) ChatStream(ctx context.Context, userID int64, source, role, message string, enableRAG bool, chunkCb func(string, bool)) error {
	inst, err := uc.factory.GetOrCreate(ctx, source, role)
	if err != nil {
		return fmt.Errorf("get agent failed: %w", err)
	}

	// 加载会话历史并传递给 Agent
	history, _ := uc.loadHistory(ctx, source, userID)
	historyMessages := uc.sessionToBladeMessages(history)

	// 注入用户偏好作为系统消息
	var instruction *blades.Message
	if prefText := uc.loadPreferences(ctx, source, userID); prefText != "" {
		instruction = blades.SystemMessage("以下是该用户的偏好信息，请遵守：\n" + prefText)
	}

	// 将 userID/source/role 放入 context，供工具函数使用
	ctx = context.WithValue(ctx, userIDContextKey, userID)
	ctx = context.WithValue(ctx, sourceContextKey, source)
	ctx = context.WithValue(ctx, roleContextKey, role)

	var fullResponse string
	err = inst.ChatStream(ctx, message, enableRAG, historyMessages, instruction, func(chunk string, done bool) {
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
