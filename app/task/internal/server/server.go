package server

import (
	"cyi-app/app/task/internal/conf"
	consul "cyi-app/pkg/kratos/config"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/google/wire"

	ts "cyi-app/pkg/kratos/server/transporter"

	aqhelper "cyi-app/pkg/util/asynq_helper"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
	mid "cyi-app/pkg/kratos/server/middleware"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewRegistrar,
	NewHTTPServerConf,
	NewHTTPServer,
	ProcessTask,
	NewServerMux,
	NewTaskServer,
	NewDiscovery,
)

func NewHTTPServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	metric *mid.Meter,
	logger log.Logger) *ts.HttpServer {
	srv := http.NewServer(ts.NewHttpServerOpt(
		&ts.TransporterOpt{
			Network: c.Http.Network,
			Addr:    c.Http.Addr,
			Timeout: c.Http.Timeout,
		},
		&ts.MiddlewareOpt{
			Logger:         logger,
			TracerProvider: tp,
			Meter:          metric,
		})...)
	return &ts.HttpServer{HttpServer: srv}
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

// 发现服务器
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

func NewTaskServer(r redis.UniversalClient, log log.Logger) *asynq.Server {
	server := aqhelper.NewServer(r, 10, log)
	return server.Server
}

func NewServerMux(log log.Logger) *ts.TaskServer {
	mux := asynq.NewServeMux()
	// 后期加个中间件
	// mux.Use()
	return &ts.TaskServer{TaskServer: mux}
}
