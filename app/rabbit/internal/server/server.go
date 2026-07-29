package server

import (
	"cyi-app/app/rabbit/internal/conf"
	kserver "cyi-app/pkg/kratos/server"
	"cyi-app/pkg/rmq"

	"github.com/go-kratos/kratos/v2/registry"

	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewRegistrar, NewQueueFactory)

// func NewGRPCServerConf(
// 	c *conf.Server,
// 	tp trace.TracerProvider,
// 	logger log.Logger,
// 	ac *conf.Auth) *ts.GrpcServer {
// 	srv := grpc.NewServer(ts.NewGrpcServerOpt(
// 		&ts.TransporterOpt{
// 			Network: c.Grpc.Network,
// 			Addr:    c.Grpc.Addr,
// 			Timeout: c.Grpc.Timeout,
// 		},
// 		&ts.MiddlewareOpt{
// 			TracerProvider: tp,
// 			Logger:         logger,
// 			ApiKey:         ac.ApiKey,
// 		})...)
// 	return &ts.GrpcServer{GrpcServer: srv}
// }

// 注册服务器
func NewRegistrar(c *conf.Registry) registry.Registrar {
	return kserver.NewRegistrar(c.Consul.Address, c.Consul.Scheme)
}

func NewQueueFactory(conn *rmq.Connection) *rmq.QueueFactory {
	return rmq.NewQueueFactory(conn)
}
