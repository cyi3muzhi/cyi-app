package server

import (
	"cyi-app/app/service/order/internal/conf"
	consul "cyi-app/pkg/kratos/config"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	ts "cyi-app/pkg/kratos/server/transporter"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewServer, NewRegistrar, NewDiscovery, NewGRPCServerConf, NewHTTPServerConf, NewHTTPServer)

// NewHTTPServerConf 创建 DTM 分支回调 HTTP 服务器
// 仅用于 DTM TCC 分支回调，不对外暴露业务接口；不配置 ApiKey 鉴权（内网隔离）
// 监听地址/超时取自 Consul 配置 server.http.*（需在 Consul 中为服务配置该段）
func NewHTTPServerConf(c *conf.Server, tp trace.TracerProvider, logger log.Logger, metric *mid.Meter) *ts.HttpServer {
	httpOpt := &ts.TransporterOpt{
		Network: c.Http.Network,
		Addr:    c.Http.Addr,
		Timeout: c.Http.Timeout,
	}
	midOpt := &ts.MiddlewareOpt{
		TracerProvider: tp,
		Logger:         logger,
		Meter:          metric,
	}
	srv := http.NewServer(ts.NewHttpServerOpt(httpOpt, midOpt)...)
	srv.Handle("/metrics", promhttp.Handler())
	return &ts.HttpServer{HttpServer: srv}
}

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

// NewRegistrar 创建 Consul 服务注册器
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

// NewDiscovery 创建 Consul 服务发现器
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
