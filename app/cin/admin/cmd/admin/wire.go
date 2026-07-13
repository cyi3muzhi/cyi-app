//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"cyi-app/app/cin/admin/internal/biz"
	"cyi-app/app/cin/admin/internal/conf"
	"cyi-app/app/cin/admin/internal/data"
	"cyi-app/app/cin/admin/internal/server"
	"cyi-app/app/cin/admin/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	mid "cyi-app/pkg/kratos/server/middleware"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Auth, *conf.Registry, log.Logger, *tracesdk.TracerProvider, *mid.Meter) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp, provideTracerProvider))
}
