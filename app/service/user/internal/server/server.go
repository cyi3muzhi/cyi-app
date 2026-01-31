package server

import (
	"cyi-app/app/service/user/internal/conf"
	consul "cyi-app/pkg/util/config"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/google/wire"

	ts "cyi-app/pkg/util/server/transporter"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewServer, NewRegistrar, NewGRPCServerConf)

func NewGRPCServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	ac *conf.Auth) *ts.GrpcServer {
	srv := grpc.NewServer(ts.NewGrpcServerOpt(
		&ts.TransporterOpt{
			Network: c.Grpc.Network,
			Addr:    c.Grpc.Addr,
			Timeout: c.Grpc.Timeout,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			Logger:         logger,
			ApiKey:         ac.ApiKey,
		})...)
	return &ts.GrpcServer{GrpcServer: srv}
}

// 注册服务器
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	reg, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address: conf.Consul.Address,
		Scheme:  conf.Consul.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return reg
}
