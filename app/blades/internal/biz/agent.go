package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	pkgblades "cyi-app/pkg/blades"

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
	factory      *AgentFactory
	rdb          redis.UniversalClient
	sessionStore *pkgblades.SessionStore // 公共会话存储
	localStor    sync.Map                // 偏好存储（Redis 不可用时 fallback）
	log          *log.Helper
}

// NewAgentUseCase 创建 AgentUseCase
func NewAgentUseCase(rdb redis.UniversalClient, factory *AgentFactory, logger log.Logger) *AgentUseCase {
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

	// 创建公共会话存储，使用 "session:" 前缀以兼容现有 key 模式
	sessionStore := pkgblades.NewSessionStore(rdb,
		pkgblades.WithPrefix(agentSessionPrefix),
		pkgblades.WithTTL(2*time.Hour),
		pkgblades.WithMaxRounds(20),
	)

	return &AgentUseCase{
		factory:      factory,
		rdb:          rdb,
		sessionStore: sessionStore,
		log:          log.NewHelper(logger),
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

// sessionKey 构造会话存储使用的 key
func sessionKey(source string, userID int64) string {
	return source + ":" + strconv.FormatInt(userID, 10)
}

// Chat 非流式聊天
func (uc *AgentUseCase) Chat(ctx context.Context, userID int64, source, role, message string, enableRAG bool) (string, error) {
	inst, err := uc.factory.GetOrCreate(ctx, source, role)
	if err != nil {
		return "", fmt.Errorf("get agent failed: %w", err)
	}

	// 使用公共会话存储加载历史并转为 blades.Message 格式
	key := sessionKey(source, userID)
	historyMessages, err := uc.sessionStore.LoadHistory(ctx, key)
	if err != nil {
		uc.log.Warnf("load history failed: %v", err)
	}

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

	uc.sessionStore.SaveHistory(ctx, key, message, response)
	return response, nil
}

// ChatStream 流式聊天
func (uc *AgentUseCase) ChatStream(ctx context.Context, userID int64, source, role, message string, enableRAG bool, chunkCb func(string, bool)) error {
	inst, err := uc.factory.GetOrCreate(ctx, source, role)
	if err != nil {
		return fmt.Errorf("get agent failed: %w", err)
	}

	// 使用公共会话存储加载历史并转为 blades.Message 格式
	key := sessionKey(source, userID)
	historyMessages, err := uc.sessionStore.LoadHistory(ctx, key)
	if err != nil {
		uc.log.Warnf("load history failed: %v", err)
	}

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

	uc.sessionStore.SaveHistory(ctx, key, message, fullResponse)
	return nil
}

// ClearHistory 清除聊天历史
func (uc *AgentUseCase) ClearHistory(ctx context.Context, userID int64, source string) error {
	key := sessionKey(source, userID)
	return uc.sessionStore.ClearHistory(ctx, key)
}

// GenerateText 生成文本（不保存聊天历史），用于 AI 辅助编写等场景
func (uc *AgentUseCase) GenerateText(ctx context.Context, prompt string) (string, error) {
	inst, err := uc.factory.GetOrCreate(ctx, "admin", "super_admin")
	if err != nil {
		return "", fmt.Errorf("get agent failed: %w", err)
	}
	instruction := blades.SystemMessage("你是一个专业的内容创作助手。请根据用户的要求生成文本内容，直接输出结果，不要有多余的解释。")
	response, err := inst.Chat(ctx, prompt, false, nil, instruction)
	if err != nil {
		return "", err
	}
	return response, nil
}

// HistoryMessage 历史消息
type HistoryMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// LoadChatHistory 加载聊天历史
func (uc *AgentUseCase) LoadChatHistory(ctx context.Context, userID int64, source string) ([]*HistoryMessage, error) {
	key := sessionKey(source, userID)
	msgs, err := uc.sessionStore.LoadRawHistory(ctx, key)
	if err != nil {
		return nil, err
	}
	history := make([]*HistoryMessage, 0, len(msgs))
	for _, m := range msgs {
		history = append(history, &HistoryMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}
	return history, nil
}
