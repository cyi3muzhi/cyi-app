package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// mockClientTransport 模拟客户端调用的 transport 上下文
type mockClientTransport struct {
	operation string
}

func (m *mockClientTransport) Kind() transport.Kind              { return transport.KindGRPC }
func (m *mockClientTransport) Endpoint() string                  { return "mock" }
func (m *mockClientTransport) Operation() string                 { return m.operation }
func (m *mockClientTransport) RequestHeader() transport.Header   { return nil }
func (m *mockClientTransport) ReplyHeader() transport.Header     { return nil }

// TestFallbackMiddleware_Success 正常调用不触发降级
func TestFallbackMiddleware_Success(t *testing.T) {
	md := fallbackMiddleware(map[string]FallbackFunc{
		"/user.v1.User/GetUser": func(_ context.Context, _ interface{}) (interface{}, error) {
			return "fallback", nil
		},
	})
	ctx := transport.NewClientContext(context.Background(), &mockClientTransport{operation: "/user.v1.User/GetUser"})
	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return "ok", nil
	}
	reply, err := md(middleware.Handler(handler))(ctx, nil)
	if err != nil {
		t.Fatalf("期望无错误，实际: %v", err)
	}
	if reply != "ok" {
		t.Fatalf("期望返回正常结果 ok，实际: %v", reply)
	}
}

// TestFallbackMiddleware_Triggered 下游失败且命中降级策略时返回兜底响应
func TestFallbackMiddleware_Triggered(t *testing.T) {
	md := fallbackMiddleware(map[string]FallbackFunc{
		"/user.v1.User/GetUser": func(_ context.Context, _ interface{}) (interface{}, error) {
			return "fallback", nil
		},
	})
	ctx := transport.NewClientContext(context.Background(), &mockClientTransport{operation: "/user.v1.User/GetUser"})
	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, errors.New("下游服务不可用")
	}
	reply, err := md(middleware.Handler(handler))(ctx, nil)
	if err != nil {
		t.Fatalf("降级后期望无错误，实际: %v", err)
	}
	if reply != "fallback" {
		t.Fatalf("期望返回兜底结果 fallback，实际: %v", reply)
	}
}

// TestFallbackMiddleware_NoPolicy 下游失败但未注册降级策略时透传错误
func TestFallbackMiddleware_NoPolicy(t *testing.T) {
	md := fallbackMiddleware(map[string]FallbackFunc{
		"/product.v1.Product/GetProduct": func(_ context.Context, _ interface{}) (interface{}, error) {
			return "fallback", nil
		},
	})
	ctx := transport.NewClientContext(context.Background(), &mockClientTransport{operation: "/user.v1.User/GetUser"})
	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, errors.New("下游服务不可用")
	}
	_, err := md(middleware.Handler(handler))(ctx, nil)
	if err == nil {
		t.Fatal("未命中降级策略时应透传错误")
	}
}
