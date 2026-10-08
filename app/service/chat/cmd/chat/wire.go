//go:build wireinject
// +build wireinject

package main

import (
	"cyi-app/app/service/chat/internal/biz"
	"cyi-app/app/service/chat/internal/conf"
	"cyi-app/app/service/chat/internal/data"
	"cyi-app/app/service/chat/internal/server"
	"cyi-app/app/service/chat/internal/service"

	mid "cyi-app/pkg/kratos/server/middleware"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	trace "go.opentelemetry.io/otel/trace"
)

func provideTracerProvider(tp *tracesdk.TracerProvider) trace.TracerProvider {
	return tp
}

func wireApp(*conf.Server, *conf.Data, *conf.Auth, *conf.Registry, log.Logger, *tracesdk.TracerProvider, *mid.Meter) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp, provideTracerProvider))
}
