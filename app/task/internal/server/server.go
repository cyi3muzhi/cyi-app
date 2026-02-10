package server

import (
	"cyi-app/app/task/internal/conf"
	consul "cyi-app/pkg/kratos/config"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/google/wire"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewServer, NewRegistrar, NewGRPCServerConf, NewHTTPServerConf, NewHTTPServer)

func NewHTTPServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	auth *conf.Auth) *ts.HttpServer {
	srv := http.NewServer(ts.NewHttpServerOpt(
		&ts.TransporterOpt{
			Network: c.Http.Network,
			Addr:    c.Http.Addr,
			Timeout: c.Http.Timeout,
		},
		&ts.MiddlewareOpt{
			ApiKey:         auth.ApiKey,
			Logger:         logger,
			TracerProvider: tp,
		})...)
	return &ts.HttpServer{HttpServer: srv}
}

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
