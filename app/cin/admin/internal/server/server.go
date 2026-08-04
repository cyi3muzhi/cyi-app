package server

import (
	"net"

	"cyi-app/app/cin/admin/internal/conf"
	consul "cyi-app/pkg/kratos/config"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"

	mid "cyi-app/pkg/kratos/server/middleware"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar, NewDiscovery, NewGRPCServerConf, NewHTTPServerConf)

// getLocalIP 获取当前IP地址
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// NewHTTPServerConf 创建 HTTP 服务器配置
func NewHTTPServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	metric *mid.Meter,
	auth *conf.Auth,
	tokenValidator mid.TokenValidator,
	enforcer mid.EnforcerProvider,
	operationLogRecorder mid.OperationLogRecorder) *ts.HttpServer {

	// 附加中间件（合并进服务中间件链末尾执行；不能再用 http.Middleware 追加，matcher.Use 会覆盖整条链）
	var extraMiddlewares []middleware.Middleware
	if operationLogRecorder != nil {
		extraMiddlewares = append(extraMiddlewares, mid.OperationLog(operationLogRecorder, logger))
	}

	opts := ts.NewHttpServerOpt(
		&ts.TransporterOpt{
			Network: c.Http.Network,
			Addr:    c.Http.Addr,
			Timeout: c.Http.Timeout,
		},
		&ts.MiddlewareOpt{
			ApiKey:           auth.ApiKey,
			WhiteList:        auth.WhiteList,
			TokenValidator:   tokenValidator,
			Enforcer:         enforcer,
			Logger:           logger,
			TracerProvider:   tp,
			Meter:            metric,
			ExtraMiddlewares: extraMiddlewares,
		})
	srv := http.NewServer(opts...)
	srv.Handle("/metrics", promhttp.Handler())
	log.NewHelper(logger).Infof("Current IP: %s", getLocalIP())
	return &ts.HttpServer{HttpServer: srv}
}

// NewGRPCServerConf 创建 gRPC 服务器配置
func NewGRPCServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	metric *mid.Meter,
	ac *conf.Auth,
	enforcer mid.EnforcerProvider,
) *ts.GrpcServer {
	srv := grpc.NewServer(ts.NewGrpcServerOpt(&ts.TransporterOpt{
		Network: c.Grpc.Network,
		Addr:    c.Grpc.Addr,
		Timeout: c.Grpc.Timeout,
	}, &ts.MiddlewareOpt{
		TracerProvider: tp,
		Logger:         logger,
		ApiKey:         ac.ApiKey,
		Enforcer:       enforcer,
		Meter:          metric,
	})...)

	return &ts.GrpcServer{GrpcServer: srv}
}

// NewRegistrar 创建服务注册器
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	reg, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address:           conf.Consul.Address,
		Scheme:            conf.Consul.Scheme,
		EnableHealthCheck: conf.Consul.EnableHealthCheck,
		CheckInterval:     conf.Consul.CheckInterval,
	})
	if err != nil {
		panic(err)
	}
	return reg
}

// NewDiscovery 创建服务发现
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
