//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"cyi-app/app/blades/internal/biz"
	"cyi-app/app/blades/internal/conf"
	"cyi-app/app/blades/internal/service"
	"cyi-app/app/blades/internal/server"
	"cyi-app/app/blades/internal/data"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/google/wire"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	trace "go.opentelemetry.io/otel/trace"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2"
)

// 将 *tracesdk.TracerProvider 转换为 trace.TracerProvider 接口
func provideTracerProvider(tp *tracesdk.TracerProvider) trace.TracerProvider {
	return tp
}

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Auth, *conf.Blades, *conf.Registry, log.Logger,*tracesdk.TracerProvider, *mid.Meter) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp, provideTracerProvider, provideBladesConfig))
}