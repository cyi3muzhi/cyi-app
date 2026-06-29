package middleware

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// TraceIDHeader 将 TraceID 注入响应 Header
func TraceIDHeader() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			traceID := getTraceID(ctx)
			if tr, ok := transport.FromServerContext(ctx); ok {
				tr.ReplyHeader().Set("X-Trace-Id", traceID)
			}
			return handler(ctx, req)
		}
	}
}

// getTraceID 获取链路追踪ID，Jaeger 未启用时用 UUID 兜底
func getTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		id := span.SpanContext().TraceID().String()
		// 排除全零（Jaeger 未初始化时的情况）
		if id != "" && id != "00000000000000000000000000000000" {
			return id
		}
	}
	// 兜底：Jaeger 没开时生成 UUID
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
