package middleware

import (
	"context"
	"sync"

	"github.com/go-kratos/aegis/circuitbreaker"
	"github.com/go-kratos/aegis/circuitbreaker/sre"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// ErrBreakerTriggered 服务端熔断触发时返回的错误（503）
// 与 kratos 客户端熔断的 ErrNotAllowed 语义一致，便于上层统一识别
var ErrBreakerTriggered = errors.New(503, "CIRCUITBREAKER", "request failed due to circuit breaker triggered")

// breakerGroup 按 operation 分组的熔断器缓存（线程安全）
// kratos 的 internal/group 不允许外部引用，这里用 sync.Map 实现等价能力
type breakerGroup struct {
	mu       sync.Mutex
	breakers map[string]circuitbreaker.CircuitBreaker
}

// newBreakerGroup 创建熔断器分组，默认使用 aegis 的 SRE 熔断器
func newBreakerGroup() *breakerGroup {
	return &breakerGroup{breakers: make(map[string]circuitbreaker.CircuitBreaker)}
}

// Get 获取指定 key 的熔断器，不存在则创建（SRE 熔断器）
func (g *breakerGroup) Get(key string) circuitbreaker.CircuitBreaker {
	g.mu.Lock()
	defer g.mu.Unlock()
	if b, ok := g.breakers[key]; ok {
		return b
	}
	b := sre.NewBreaker()
	g.breakers[key] = b
	return b
}

// CircuitBreakerServer 服务端熔断中间件
// 按接口维度（operation，如 /order.v1.Order/CreateOrder）统计请求成功率，
// 错误率超过阈值（默认触发条件）时熔断打开：直接返回 503 快速失败，保护服务自身不被异常流量压垮。
// 熔断打开后按 SRE 算法逐步放行探测（半开），恢复后自动闭合。
func CircuitBreakerServer() middleware.Middleware {
	group := newBreakerGroup()
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			info, ok := transport.FromServerContext(ctx)
			if !ok || info.Operation() == "" {
				return handler(ctx, req)
			}
			breaker := group.Get(info.Operation())
			if err := breaker.Allow(); err != nil {
				// 熔断拒绝：继续计数让熔断比例更高
				breaker.MarkFailed()
				return nil, ErrBreakerTriggered
			}
			reply, err := handler(ctx, req)
			// 仅对真正的服务端错误（5xx）计入失败，业务参数错误（4xx）不影响熔断
			if err != nil && (errors.IsInternalServer(err) || errors.IsServiceUnavailable(err) || errors.IsGatewayTimeout(err)) {
				breaker.MarkFailed()
			} else {
				breaker.MarkSuccess()
			}
			return reply, err
		}
	}
}
