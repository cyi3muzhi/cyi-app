//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/conf"
	"cyi-app/app/service/order/internal/data"
	"cyi-app/app/service/order/internal/server"
	"cyi-app/app/service/order/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	trace "go.opentelemetry.io/otel/trace"
	mid "cyi-app/pkg/kratos/server/middleware"
)

// 将 *tracesdk.TracerProvider 转换为 trace.TracerProvider 接口
func provideTracerProvider(tp *tracesdk.TracerProvider) trace.TracerProvider {
	return tp
}

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Auth, *conf.Registry, log.Logger,*tracesdk.TracerProvider, *mid.Meter) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp, provideTracerProvider))
}
