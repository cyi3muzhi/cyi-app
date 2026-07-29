package blades

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-kratos/blades"
	"github.com/redis/go-redis/v9"
)

// setupTestRedis 设置测试用的Redis
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, redis.UniversalClient) {
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

// TestLoadHistory_EmptySessionID 测试空sessionID返回nil
func TestLoadHistory_EmptySessionID(t *testing.T) {
	_, rdb := setupTestRedis(t)
	ctx := context.Background()

	history, err := LoadHistory(ctx, rdb, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if history != nil {
		t.Errorf("expected nil history for empty sessionID, got %v", history)
	}
}

// TestLoadHistory_NoSession 测试不存在的会话返回空列表
func TestLoadHistory_NoSession(t *testing.T) {
	_, rdb := setupTestRedis(t)
	ctx := context.Background()

	history, err := LoadHistory(ctx, rdb, "nonexistent-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if history != nil {
		t.Errorf("expected nil history for nonexistent session, got %v", history)
	}
	if len(history) != 0 {
		t.Errorf("expected empty history, got length %d", len(history))
	}
}

// TestSaveHistory_EmptySessionID 测试空sessionID不保存
func TestSaveHistory_EmptySessionID(t *testing.T) {
	_, rdb := setupTestRedis(t)
	ctx := context.Background()

	err := SaveHistory(ctx, rdb, "", "hello", "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSaveAndLoadHistory 测试保存和加载会话历史
func TestSaveAndLoadHistory(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 保存一轮对话
	err := SaveHistory(ctx, rdb, sessionID, "你好", "你好！有什么可以帮助你的？")
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	// 验证Redis中存在数据
	key := defaultSessionPrefix + sessionID
	if !mr.Exists(key) {
		t.Error("expected key to exist in Redis")
	}

	// 加载历史
	history, err := LoadHistory(ctx, rdb, sessionID)
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(history))
	}

	// 验证用户消息
	if history[0].Role != blades.RoleUser {
		t.Errorf("expected first message role to be user, got %v", history[0].Role)
	}
	if history[0].Text() != "你好" {
		t.Errorf("expected user message '你好', got '%s'", history[0].Text())
	}

	// 验证助手消息
	if history[1].Role != blades.RoleAssistant {
		t.Errorf("expected second message role to be assistant, got %v", history[1].Role)
	}
	if history[1].Text() != "你好！有什么可以帮助你的？" {
		t.Errorf("expected assistant message '你好！有什么可以帮助你的？', got '%s'", history[1].Text())
	}
	if history[1].Status != blades.StatusCompleted {
		t.Errorf("expected assistant message status to be completed, got %v", history[1].Status)
	}
}

// TestSaveHistory_MultipleRounds 测试保存多轮对话
func TestSaveHistory_MultipleRounds(t *testing.T) {
	_, rdb := setupTestRedis(t)
	ctx := context.Background()
	sessionID := "test-session-multi"

	// 保存3轮对话
	for i := 1; i <= 3; i++ {
		userMsg := fmt.Sprintf("用户问题%d", i)
		assistantMsg := fmt.Sprintf("AI回答%d", i)
		err := SaveHistory(ctx, rdb, sessionID, userMsg, assistantMsg)
		if err != nil {
			t.Fatalf("SaveHistory round %d failed: %v", i, err)
		}
	}

	// 加载历史
	history, err := LoadHistory(ctx, rdb, sessionID)
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	// 3轮对话 = 6条消息（每轮2条）
	if len(history) != 6 {
		t.Fatalf("expected 6 messages, got %d", len(history))
	}

	// 验证第一轮
	if history[0].Text() != "用户问题1" {
		t.Errorf("expected first user message '用户问题1', got '%s'", history[0].Text())
	}
	if history[1].Text() != "AI回答1" {
		t.Errorf("expected first assistant message 'AI回答1', got '%s'", history[1].Text())
	}

	// 验证最后一轮
	if history[4].Text() != "用户问题3" {
		t.Errorf("expected last user message '用户问题3', got '%s'", history[4].Text())
	}
	if history[5].Text() != "AI回答3" {
		t.Errorf("expected last assistant message 'AI回答3', got '%s'", history[5].Text())
	}
}

// TestSaveHistory_Truncation 测试历史记录截断（超过20轮）
func TestSaveHistory_Truncation(t *testing.T) {
	_, rdb := setupTestRedis(t)
	ctx := context.Background()
	sessionID := "test-session-truncate"

	// 保存25轮对话（超过20轮限制）
	for i := 1; i <= 25; i++ {
		userMsg := fmt.Sprintf("Q%d", i)
		assistantMsg := fmt.Sprintf("A%d", i)
		err := SaveHistory(ctx, rdb, sessionID, userMsg, assistantMsg)
		if err != nil {
			t.Fatalf("SaveHistory round %d failed: %v", i, err)
		}
	}

	// 加载历史
	history, err := LoadHistory(ctx, rdb, sessionID)
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	// 应该只保留最近20轮 = 40条消息
	if len(history) != 40 {
		t.Fatalf("expected 40 messages (20 rounds), got %d", len(history))
	}

	// 第一条应该是第6轮的问题（前5轮被截断）
	if history[0].Text() != "Q6" {
		t.Errorf("expected first message to be 'Q6', got '%s'", history[0].Text())
	}

	// 最后一条应该是第25轮的回答
	if history[39].Text() != "A25" {
		t.Errorf("expected last message to be 'A25', got '%s'", history[39].Text())
	}
}

// TestClearHistory_EmptySessionID 测试清空空sessionID
func TestClearHistory_EmptySessionID(t *testing.T) {
	_, rdb := setupTestRedis(t)
	ctx := context.Background()

	err := ClearHistory(ctx, rdb, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClearHistory 测试清空会话历史
func TestClearHistory(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	ctx := context.Background()
	sessionID := "test-session-clear"

	// 先保存一些数据
	err := SaveHistory(ctx, rdb, sessionID, "hello", "hi")
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	key := defaultSessionPrefix + sessionID
	if !mr.Exists(key) {
		t.Fatal("expected key to exist before clear")
	}

	// 清空历史
	err = ClearHistory(ctx, rdb, sessionID)
	if err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}

	// 验证key已被删除
	if mr.Exists(key) {
		t.Error("expected key to be deleted after clear")
	}

	// 加载应该返回空
	history, err := LoadHistory(ctx, rdb, sessionID)
	if err != nil {
		t.Fatalf("LoadHistory after clear failed: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("expected empty history after clear, got %d messages", len(history))
	}
}

// TestLoadHistory_InvalidJSON 测试加载损坏的JSON数据
func TestLoadHistory_InvalidJSON(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	ctx := context.Background()
	sessionID := "test-session-bad-json"

	// 手动写入无效的JSON
	key := defaultSessionPrefix + sessionID
	mr.Set(key, "invalid json data")

	_, err := LoadHistory(ctx, rdb, sessionID)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestLoadHistory_UnknownRole 测试加载包含未知角色的消息
func TestLoadHistory_UnknownRole(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	ctx := context.Background()
	sessionID := "test-session-unknown-role"

	// 手动写入包含未知角色的JSON
	key := defaultSessionPrefix + sessionID
	badData := `[{"role":"system","content":"you are helpful"},{"role":"user","content":"hello"},{"role":"assistant","content":"hi"}]`
	mr.Set(key, badData)

	history, err := LoadHistory(ctx, rdb, sessionID)
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	// 应该只有2条消息（user和assistant），system角色被忽略
	if len(history) != 2 {
		t.Errorf("expected 2 messages (user+assistant), got %d", len(history))
	}
}

// TestSaveHistory_TTL 测试会话历史的TTL设置
func TestSaveHistory_TTL(t *testing.T) {
	mr, rdb := setupTestRedis(t)
	ctx := context.Background()
	sessionID := "test-session-ttl"

	// 保存历史
	err := SaveHistory(ctx, rdb, sessionID, "hello", "hi")
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	key := defaultSessionPrefix + sessionID
	ttl := mr.TTL(key)

	// TTL应该接近2小时（7200秒）
	// 由于执行时间，可能稍微少一点，但应该大于7000秒
	if ttl < 7000 {
		t.Errorf("expected TTL > 7000s (approx 2h), got %v", ttl)
	}
}

// TestSessionIsolation 测试不同会话之间的隔离
func TestSessionIsolation(t *testing.T) {
	_, rdb := setupTestRedis(t)
	ctx := context.Background()

	// 保存会话1
	err := SaveHistory(ctx, rdb, "session1", "msg1", "reply1")
	if err != nil {
		t.Fatalf("SaveHistory session1 failed: %v", err)
	}

	// 保存会话2
	err = SaveHistory(ctx, rdb, "session2", "msg2", "reply2")
	if err != nil {
		t.Fatalf("SaveHistory session2 failed: %v", err)
	}

	// 加载会话1
	history1, err := LoadHistory(ctx, rdb, "session1")
	if err != nil {
		t.Fatalf("LoadHistory session1 failed: %v", err)
	}
	if len(history1) != 2 || history1[0].Text() != "msg1" {
		t.Errorf("session1 history mismatch")
	}

	// 加载会话2
	history2, err := LoadHistory(ctx, rdb, "session2")
	if err != nil {
		t.Fatalf("LoadHistory session2 failed: %v", err)
	}
	if len(history2) != 2 || history2[0].Text() != "msg2" {
		t.Errorf("session2 history mismatch")
	}

	// 清空会话1不影响会话2
	err = ClearHistory(ctx, rdb, "session1")
	if err != nil {
		t.Fatalf("ClearHistory session1 failed: %v", err)
	}

	history2After, err := LoadHistory(ctx, rdb, "session2")
	if err != nil {
		t.Fatalf("LoadHistory session2 after clear session1 failed: %v", err)
	}
	if len(history2After) != 2 {
		t.Errorf("session2 should not be affected by clearing session1, got %d messages", len(history2After))
	}
}
