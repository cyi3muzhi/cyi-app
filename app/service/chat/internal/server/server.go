package server

import (
	"cyi-app/app/service/chat/internal/conf"
	consul "cyi-app/pkg/kratos/config"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/google/wire"

	ts "cyi-app/pkg/kratos/server/transporter"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
)

var ProviderSet = wire.NewSet(NewServer, NewRegistrar, NewDiscovery, NewGRPCServerConf)

// NewGRPCServerConf 创建 gRPC 服务器（chat 服务仅内部 gRPC，不提供 HTTP）
func NewGRPCServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	metric *mid.Meter,
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
			Meter:          metric,
		})...)
	return &ts.GrpcServer{GrpcServer: srv}
}

// NewRegistrar 创建 Consul 注册器
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

// NewDiscovery 创建 Consul 服务发现
func NewDiscovery(conf *conf.Registry) registry.Discovery {
	dis, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address: conf.Consul.Address,
		Scheme:  conf.Consul.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return dis
}
