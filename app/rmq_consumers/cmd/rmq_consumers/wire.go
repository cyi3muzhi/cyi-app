//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"cyi-app/app/rmq_consumers/internal/biz"
	"cyi-app/app/rmq_consumers/internal/conf"
	"cyi-app/app/rmq_consumers/internal/consumers"
	"cyi-app/app/rmq_consumers/internal/data"
	"cyi-app/app/rmq_consumers/internal/server"
	"cyi-app/app/rmq_consumers/internal/service"

	"cyi-app/pkg/rmq"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)


// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Auth, *conf.Registry, log.Logger,*rmq.Connection) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, consumers.ProviderSet, newApp))
}
