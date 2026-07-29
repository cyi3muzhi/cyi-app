package biz

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/kratos/v2/log"
)

// AgentLogging 日志中间件：记录 Agent 调用日志（含耗时）
func AgentLogging(logger log.Logger) blades.Middleware {
	return func(next blades.Handler) blades.Handler {
		return blades.HandleFunc(func(ctx context.Context, invocation *blades.Invocation) blades.Generator[*blades.Message, error] {
			start := time.Now()
			helper := log.NewHelper(logger)
			helper.Infof("agent request: model=%s, message=%s", invocation.Model, truncate(invocation.Message.Text(), 100))

			return func(yield func(*blades.Message, error) bool) {
				stream := next.Handle(ctx, invocation)
				var lastErr error
				for msg, err := range stream {
					if err != nil {
						lastErr = err
					}
					if !yield(msg, err) {
						break
					}
				}
				elapsed := time.Since(start)
				if lastErr != nil {
					helper.Errorf("agent failed after %s: %v", elapsed, lastErr)
				} else {
					helper.Infof("agent success after %s", elapsed)
				}
			}
		})
	}
}

// AgentRecovery 恢复中间件：捕获 Agent 执行中的 panic
func AgentRecovery(logger log.Logger) blades.Middleware {
	return func(next blades.Handler) blades.Handler {
		return blades.HandleFunc(func(ctx context.Context, invocation *blades.Invocation) blades.Generator[*blades.Message, error] {
			return func(yield func(*blades.Message, error) bool) {
				defer func() {
					if r := recover(); r != nil {
						stack := string(debug.Stack())
						log.NewHelper(logger).Errorf("agent panic recovered: %v\nstack: %s", r, stack)
						yield(nil, fmt.Errorf("agent panic: %v", r))
					}
				}()
				stream := next.Handle(ctx, invocation)
				for msg, err := range stream {
					if !yield(msg, err) {
						break
					}
				}
			}
		})
	}
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
