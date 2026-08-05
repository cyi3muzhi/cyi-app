package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	preferenceTTL      = 30 * 24 * time.Hour // 偏好数据 30 天过期，用户活跃时自动续期
	cleanupInterval    = 1 * time.Hour       // 内存清理间隔
	rateLimitWindow    = 100 * time.Millisecond // 速率限制窗口（每秒最多 10 次）
)

// contextKey 用于 context.WithValue 的 key 类型
type contextKey string

const (
	userIDContextKey contextKey = "user_id"
	sourceContextKey contextKey = "source"
	roleContextKey   contextKey = "role"
)

// rateEntry 速率限制条目
type rateEntry struct {
	mu       sync.Mutex
	lastTime time.Time
}

// prefStorage 本地偏好存储（Redis 不可用时的 fallback，并发安全）
type prefStorage struct {
	mu   sync.Mutex
	data map[string]map[string]string
}

// newPrefStorage 创建本地偏好存储
func newPrefStorage() *prefStorage {
	return &prefStorage{data: make(map[string]map[string]string)}
}

// load 加载用户偏好（返回副本，避免外部修改影响内部数据）
func (p *prefStorage) load(key string) map[string]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	src := p.data[key]
	if src == nil {
		return nil
	}
	cp := make(map[string]string, len(src))
	for k, v := range src {
		cp[k] = v
	}
	return cp
}

// save 保存用户偏好（深拷贝写入，避免外部引用逃逸）
func (p *prefStorage) save(key string, prefs map[string]string) {
	cp := make(map[string]string, len(prefs))
	for k, v := range prefs {
		cp[k] = v
	}
	p.mu.Lock()
	p.data[key] = cp
	p.mu.Unlock()
}

// AgentUseCase 智能体用例（组合 AgentFactory + 会话管理）
type AgentUseCase struct {
	factory      *AgentFactory
	rdb          redis.UniversalClient
	sessionStore *pkgblades.SessionStore // 公共会话存储
	localPrefs   *prefStorage            // 偏好存储（Redis 不可用时 fallback）
	log          *log.Helper
	sessionMu    sync.Map // 每用户会话互斥锁，key 为 source:userID
	rateLimit    sync.Map // 速率限制，key 为 source:userID
	stopCleanup  chan struct{}
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

	uc := &AgentUseCase{
		factory:      factory,
		rdb:          rdb,
		sessionStore: sessionStore,
		localPrefs:   newPrefStorage(),
		log:          log.NewHelper(logger),
		stopCleanup:  make(chan struct{}),
	}

	// 启动后台 goroutine 定期清理无活跃用户的 sessionMu 和 rateLimit 条目
	go uc.startCleanupLoop()

	return uc
}

// startCleanupLoop 定期清理内存中的无活跃条目，防止内存泄漏
func (uc *AgentUseCase) startCleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	lastActive := make(map[string]time.Time) // 记录每个 key 的最后活跃时间

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			// 清理超过 2 小时未活跃的 sessionMu 条目
			uc.sessionMu.Range(func(key, val any) bool {
				k := key.(string)
				last, ok := lastActive[k]
				if !ok {
					lastActive[k] = now
					return true
				}
				if now.Sub(last) > 2*time.Hour {
					uc.sessionMu.Delete(key)
					delete(lastActive, k)
				}
				return true
			})
			// rateLimit 条目也做类似清理（超过 1 小时未活跃的删除）
			uc.rateLimit.Range(func(key, val any) bool {
				entry := val.(*rateEntry)
				entry.mu.Lock()
				idle := now.Sub(entry.lastTime)
				entry.mu.Unlock()
				if idle > time.Hour {
					uc.rateLimit.Delete(key)
				}
				return true
			})
		case <-uc.stopCleanup:
			return
		}
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
			if e := json.Unmarshal(data, &prefs); e != nil {
				uc.log.Warnf("unmarshal preference failed: %v", e)
			}
		}
	} else {
		prefs = uc.localPrefs.load(key)
	}

	if len(prefs) == 0 {
		return ""
	}

	var sb strings.Builder
	for k, v := range prefs {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
	}
	return sb.String()
}

// savePreference 保存单条用户偏好
func (uc *AgentUseCase) savePreference(ctx context.Context, source string, userID int64, key_, value string) error {
	key := preferenceKey(source, userID)

	var prefs map[string]string
	if uc.rdb != nil {
		data, err := uc.rdb.Get(ctx, key).Bytes()
		if err == nil {
			if e := json.Unmarshal(data, &prefs); e != nil {
				uc.log.Warnf("unmarshal preference failed: %v", e)
			}
		}
	} else {
		prefs = uc.localPrefs.load(key)
	}

	if prefs == nil {
		prefs = make(map[string]string)
	}
	prefs[key_] = value

	payload, _ := json.Marshal(prefs)
	if uc.rdb != nil {
		// 设置 TTL，用户活跃时自动续期
		return uc.rdb.Set(ctx, key, payload, preferenceTTL).Err()
	}
	uc.localPrefs.save(key, prefs)
	return nil
}

// sessionKey 构造会话存储使用的 key
func sessionKey(source string, userID int64) string {
	return source + ":" + strconv.FormatInt(userID, 10)
}

// checkRateLimit 检查用户请求是否超过速率限制（每秒最多 10 次）
func (uc *AgentUseCase) checkRateLimit(userID int64, source string) error {
	key := fmt.Sprintf("%s:%d", source, userID)
	val, _ := uc.rateLimit.LoadOrStore(key, &rateEntry{})
	entry := val.(*rateEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	now := time.Now()
	if now.Sub(entry.lastTime) < rateLimitWindow {
		return fmt.Errorf("rate limit exceeded")
	}
	entry.lastTime = now
	return nil
}

// buildInstruction 构建系统指令（合并用户偏好和 RAG 上下文）
func (uc *AgentUseCase) buildInstruction(ctx context.Context, inst *AgentInstance, source string, userID int64, message string, enableRAG bool) *blades.Message {
	var parts []string

	// 1. 用户偏好
	if prefText := uc.loadPreferences(ctx, source, userID); prefText != "" {
		parts = append(parts, "以下是该用户的偏好信息，请遵守：\n"+prefText)
	}

	// 2. RAG 检索结果（作为系统上下文，而非拼接到用户消息）
	if enableRAG && inst.rag != nil && inst.rag.ShouldRetrieve(message) {
		results, err := inst.rag.Retrieve(ctx, inst.knowledgeTag, message, 0)
		if err == nil && len(results) > 0 {
			parts = append(parts, FormatResults(results))
		}
	}

	if len(parts) == 0 {
		return nil
	}
	return blades.SystemMessage(strings.Join(parts, "\n\n"))
}

// Chat 非流式聊天
func (uc *AgentUseCase) Chat(ctx context.Context, userID int64, source, role, message string, enableRAG bool) (string, error) {
	if err := uc.checkRateLimit(userID, source); err != nil {
		return "", err
	}

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

	// 构建系统指令（用户偏好 + RAG 上下文）
	instruction := uc.buildInstruction(ctx, inst, source, userID, message, enableRAG)

	// 将 userID/source/role 放入 context，供工具函数使用
	ctx = WithUserID(ctx, userID)
	ctx = WithSource(ctx, source)
	ctx = WithRole(ctx, role)

	response, err := inst.Chat(ctx, message, instruction, historyMessages)
	if err != nil {
		return "", err
	}

	// 加锁保护并发会话写入
	muKey := key
	mu := &sync.Mutex{}
	if actual, loaded := uc.sessionMu.LoadOrStore(muKey, mu); loaded {
		mu = actual.(*sync.Mutex)
	}
	mu.Lock()
	uc.sessionStore.SaveHistory(ctx, key, message, response)
	mu.Unlock()
	return response, nil
}

// ChatStream 流式聊天
func (uc *AgentUseCase) ChatStream(ctx context.Context, userID int64, source, role, message string, enableRAG bool, chunkCb func(string, bool)) error {
	if err := uc.checkRateLimit(userID, source); err != nil {
		return err
	}

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

	// 构建系统指令（用户偏好 + RAG 上下文）
	instruction := uc.buildInstruction(ctx, inst, source, userID, message, enableRAG)

	// 将 userID/source/role 放入 context，供工具函数使用
	ctx = WithUserID(ctx, userID)
	ctx = WithSource(ctx, source)
	ctx = WithRole(ctx, role)

	var fullResponse strings.Builder
	err = inst.ChatStream(ctx, message, instruction, historyMessages, func(chunk string, done bool) {
		fullResponse.WriteString(chunk)
		chunkCb(chunk, done)
	})
	if err != nil {
		return err
	}

	// 加锁保护并发会话写入
	muKey := key
	mu := &sync.Mutex{}
	if actual, loaded := uc.sessionMu.LoadOrStore(muKey, mu); loaded {
		mu = actual.(*sync.Mutex)
	}
	mu.Lock()
	uc.sessionStore.SaveHistory(ctx, key, message, fullResponse.String())
	mu.Unlock()
	return nil
}

// ClearHistory 清除聊天历史
func (uc *AgentUseCase) ClearHistory(ctx context.Context, userID int64, source string) error {
	key := sessionKey(source, userID)
	return uc.sessionStore.ClearHistory(ctx, key)
}

// GenerateText 生成文本（不保存聊天历史），用于 AI 辅助编写等场景
func (uc *AgentUseCase) GenerateText(ctx context.Context, prompt string) (string, error) {
	// 优先从 context 提取角色和来源，如果拿不到则使用默认值
	role := GetRole(ctx)
	source := GetSource(ctx)
	if role == "" {
		role = "super_admin"
	}
	if source == "" {
		source = "admin"
	}
	inst, err := uc.factory.GetOrCreate(ctx, source, role)
	if err != nil {
		return "", fmt.Errorf("get agent failed: %w", err)
	}
	instruction := blades.SystemMessage("你是一个专业的内容创作助手。请根据用户的要求生成文本内容，直接输出结果，不要有多余的解释。")
	response, err := inst.Chat(ctx, prompt, instruction, nil)
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
