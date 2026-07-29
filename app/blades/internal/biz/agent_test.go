package biz

import (
	"context"
	"fmt"
	"testing"
	"time"

	pkgblades "cyi-app/pkg/blades"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// setupMiniredis 启动 miniredis 用于测试
func setupMiniredis(t *testing.T) (*miniredis.Miniredis, redis.UniversalClient) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(func() {
		mr.Close()
	})

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	t.Cleanup(func() {
		rdb.Close()
	})

	return mr, rdb
}

// newSessionStore 创建测试用的 SessionStore
func newSessionStore(rdb redis.UniversalClient, prefix string) *pkgblades.SessionStore {
	return pkgblades.NewSessionStore(rdb,
		pkgblades.WithPrefix(prefix),
		pkgblades.WithTTL(2*time.Hour),
		pkgblades.WithMaxRounds(20),
	)
}

// TestLoadChatHistory_Empty 测试无历史记录时返回空数组
func TestLoadChatHistory_Empty(t *testing.T) {
	_, rdb := setupMiniredis(t)

	uc := &AgentUseCase{
		rdb:          rdb,
		sessionStore: newSessionStore(rdb, "session:"),
	}

	history, err := uc.LoadChatHistory(context.Background(), 1001, "user")
	if err != nil {
		t.Fatalf("LoadChatHistory failed: %v", err)
	}
	if history == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(history) != 0 {
		t.Errorf("expected empty history, got %d items", len(history))
	}
}

// TestLoadChatHistory_WithData 测试有历史记录时正确返回
func TestLoadChatHistory_WithData(t *testing.T) {
	_, rdb := setupMiniredis(t)

	uc := &AgentUseCase{
		rdb:          rdb,
		sessionStore: newSessionStore(rdb, "session:"),
	}

	ctx := context.Background()
	key := "user:1001"
	err := uc.sessionStore.SaveHistory(ctx, key, "你好", "你好！有什么可以帮助你的？")
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	history, err := uc.LoadChatHistory(ctx, 1001, "user")
	if err != nil {
		t.Fatalf("LoadChatHistory failed: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(history))
	}
	if history[0].Role != "user" || history[0].Content != "你好" {
		t.Errorf("expected user message '你好', got role=%s content=%s", history[0].Role, history[0].Content)
	}
	if history[1].Role != "assistant" || history[1].Content != "你好！有什么可以帮助你的？" {
		t.Errorf("expected assistant message, got role=%s content=%s", history[1].Role, history[1].Content)
	}
	// 验证时间戳不为零
	if history[0].Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}
	if history[1].Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}
}

// TestLoadChatHistory_MultipleRounds 测试多轮对话
func TestLoadChatHistory_MultipleRounds(t *testing.T) {
	_, rdb := setupMiniredis(t)

	uc := &AgentUseCase{
		rdb:          rdb,
		sessionStore: newSessionStore(rdb, "session:"),
	}

	ctx := context.Background()
	key := "user:1002"

	for i := 1; i <= 3; i++ {
		userMsg := fmt.Sprintf("问题%d", i)
		assistantMsg := fmt.Sprintf("回答%d", i)
		err := uc.sessionStore.SaveHistory(ctx, key, userMsg, assistantMsg)
		if err != nil {
			t.Fatalf("SaveHistory round %d failed: %v", i, err)
		}
	}

	history, err := uc.LoadChatHistory(ctx, 1002, "user")
	if err != nil {
		t.Fatalf("LoadChatHistory failed: %v", err)
	}
	if len(history) != 6 {
		t.Fatalf("expected 6 messages (3 rounds), got %d", len(history))
	}
	if history[4].Content != "问题3" {
		t.Errorf("expected '问题3', got '%s'", history[4].Content)
	}
	if history[5].Content != "回答3" {
		t.Errorf("expected '回答3', got '%s'", history[5].Content)
	}
}

// TestClearChatHistory 测试清除历史
func TestClearChatHistory(t *testing.T) {
	mr, rdb := setupMiniredis(t)

	uc := &AgentUseCase{
		rdb:          rdb,
		sessionStore: newSessionStore(rdb, "session:"),
	}

	ctx := context.Background()
	key := "user:1003"

	err := uc.sessionStore.SaveHistory(ctx, key, "hi", "hello")
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	redisKey := "session:" + key
	if !mr.Exists(redisKey) {
		t.Fatal("expected key to exist in Redis before clear")
	}

	err = uc.ClearHistory(ctx, 1003, "user")
	if err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}

	if mr.Exists(redisKey) {
		t.Error("expected key to be deleted after clear")
	}

	history, err := uc.LoadChatHistory(ctx, 1003, "user")
	if err != nil {
		t.Fatalf("LoadChatHistory after clear failed: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("expected empty history after clear, got %d messages", len(history))
	}
}

// TestChatHistory_UserIsolation 测试不同用户/来源之间的会话隔离
func TestChatHistory_UserIsolation(t *testing.T) {
	_, rdb := setupMiniredis(t)

	uc := &AgentUseCase{
		rdb:          rdb,
		sessionStore: newSessionStore(rdb, "session:"),
	}

	ctx := context.Background()

	uc.sessionStore.SaveHistory(ctx, "user:101", "用户A消息", "回复A")
	uc.sessionStore.SaveHistory(ctx, "admin:102", "用户B消息", "回复B")

	historyA, _ := uc.LoadChatHistory(ctx, 101, "user")
	if len(historyA) != 2 || historyA[0].Content != "用户A消息" {
		t.Errorf("user A history mismatch: %+v", historyA)
	}

	historyB, _ := uc.LoadChatHistory(ctx, 102, "admin")
	if len(historyB) != 2 || historyB[0].Content != "用户B消息" {
		t.Errorf("user B history mismatch: %+v", historyB)
	}

	uc.ClearHistory(ctx, 101, "user")
	historyBAfter, _ := uc.LoadChatHistory(ctx, 102, "admin")
	if len(historyBAfter) != 2 {
		t.Errorf("user B should not be affected by clearing user A, got %d messages", len(historyBAfter))
	}
}

// TestChatHistory_MemoryFallback 测试 Redis 不可用时内存 fallback
func TestChatHistory_MemoryFallback(t *testing.T) {
	uc := &AgentUseCase{
		rdb:          nil,
		sessionStore: newSessionStore(nil, "session:"),
	}

	ctx := context.Background()

	err := uc.sessionStore.SaveHistory(ctx, "user:2001", "hello", "world")
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	history, err := uc.LoadChatHistory(ctx, 2001, "user")
	if err != nil {
		t.Fatalf("LoadChatHistory failed: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(history))
	}
	if history[0].Content != "hello" || history[1].Content != "world" {
		t.Errorf("unexpected content: %+v", history)
	}

	err = uc.ClearHistory(ctx, 2001, "user")
	if err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}

	history, err = uc.LoadChatHistory(ctx, 2001, "user")
	if err != nil {
		t.Fatalf("LoadChatHistory after clear failed: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("expected empty history after clear, got %d messages", len(history))
	}
}

// TestChatHistory_TTL 测试会话历史的 TTL 设置
func TestChatHistory_TTL(t *testing.T) {
	mr, rdb := setupMiniredis(t)

	uc := &AgentUseCase{
		rdb:          rdb,
		sessionStore: newSessionStore(rdb, "session:"),
	}

	key := "user:3001"
	err := uc.sessionStore.SaveHistory(context.Background(), key, "hello", "hi")
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	redisKey := "session:" + key
	ttl := mr.TTL(redisKey)

	if ttl < 7000 {
		t.Errorf("expected TTL > 7000s (approx 2h), got %v", ttl)
	}
}

// TestChatHistory_Truncation 测试历史记录截断
func TestChatHistory_Truncation(t *testing.T) {
	_, rdb := setupMiniredis(t)

	uc := &AgentUseCase{
		rdb:          rdb,
		sessionStore: newSessionStore(rdb, "session:"),
	}

	ctx := context.Background()
	key := "user:4001"

	for i := 1; i <= 25; i++ {
		userMsg := fmt.Sprintf("Q%d", i)
		assistantMsg := fmt.Sprintf("A%d", i)
		err := uc.sessionStore.SaveHistory(ctx, key, userMsg, assistantMsg)
		if err != nil {
			t.Fatalf("SaveHistory round %d failed: %v", i, err)
		}
	}

	history, err := uc.LoadChatHistory(ctx, 4001, "user")
	if err != nil {
		t.Fatalf("LoadChatHistory failed: %v", err)
	}

	if len(history) != 40 {
		t.Fatalf("expected 40 messages (20 rounds), got %d", len(history))
	}

	if history[0].Content != "Q6" {
		t.Errorf("expected first message to be 'Q6', got '%s'", history[0].Content)
	}

	if history[39].Content != "A25" {
		t.Errorf("expected last message to be 'A25', got '%s'", history[39].Content)
	}
}

// TestSessionStore_KeyPrefix 测试使用自定义前缀的 session key 格式
func TestSessionStore_KeyPrefix(t *testing.T) {
	mr, rdb := setupMiniredis(t)

	store := newSessionStore(rdb, "app:")
	ctx := context.Background()

	err := store.SaveHistory(ctx, "custom:123", "msg", "reply")
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	if !mr.Exists("app:custom:123") {
		t.Error("expected key 'app:custom:123' to exist")
	}
	if mr.Exists("session:custom:123") {
		t.Error("did not expect key 'session:custom:123' to exist")
	}
}
