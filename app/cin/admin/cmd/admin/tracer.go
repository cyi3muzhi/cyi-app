package main

import (
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	trace "go.opentelemetry.io/otel/trace"
)

// provideTracerProvider 将 *tracesdk.TracerProvider 转换为 trace.TracerProvider 接口
func provideTracerProvider(tp *tracesdk.TracerProvider) trace.TracerProvider {
	return tp
}
