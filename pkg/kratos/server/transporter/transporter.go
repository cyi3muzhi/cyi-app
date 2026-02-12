package transporter

import (
	"fmt"
	"os"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/durationpb"
)

type TransporterOpt struct {
	Network string
	Addr    string
	Timeout *durationpb.Duration

	// 客户端服务发现
	ServiceName string
	Discovery   registry.Discovery
}

type MiddlewareOpt struct {
	// 认证 可选
	ApiKey     string
	ServiceKey string
	WhiteList  []string

	// 链路追踪 *
	TracerProvider trace.TracerProvider
	// 日志 *
	Logger log.Logger
}

// http服务器选项
func NewHttpServerOpt(httpOpt *TransporterOpt, midOpt *MiddlewareOpt) []http.ServerOption {
	mids := []mid.MiddlewareOption{}
	if midOpt.TracerProvider == nil {
		panic("NewHttpServerOpt TracerProvider is nil")
	}
	if midOpt.Logger == nil {
		panic("NewHttpServerOpt Logger is nil")
	}
	mids = append(mids, mid.WithLoggerMiddleware(midOpt.Logger))
	if midOpt.ApiKey != "" {
		mids = append(mids, mid.WithAuthMiddleware(mid.Auth{
			ApiKey:    midOpt.ApiKey,
			WhiteList: midOpt.WhiteList,
		}))
	}
	middlewares := mid.ServerMiddlewares(midOpt.TracerProvider, mids...)

	opts := []http.ServerOption{
		http.Middleware(middlewares...),
		http.Filter(handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}),
		)),
	}
	if httpOpt.Network != "" {
		opts = append(opts, http.Network(httpOpt.Network))
	}
	if httpOpt.Addr != "" {
		opts = append(opts, http.Address(httpOpt.Addr))
	}
	if httpOpt.Timeout != nil {
		opts = append(opts, http.Timeout(httpOpt.Timeout.AsDuration()))
	}
	return opts
}

// grpc服务器选项
func NewGrpcServerOpt(grpcOpt *TransporterOpt, midOpt *MiddlewareOpt) []grpc.ServerOption {
	mids := []mid.MiddlewareOption{}
	if midOpt.TracerProvider == nil {
		panic("NewGrpcServerOpt TracerProvider is nil")
	}
	if midOpt.Logger == nil {
		panic("NewGrpcServerOpt Logger is nil")
	}
	mids = append(mids, mid.WithLoggerMiddleware(midOpt.Logger))
	if midOpt.ApiKey != "" {
		mids = append(mids, mid.WithAuthMiddleware(mid.Auth{
			ApiKey: midOpt.ApiKey,
		}))
	}
	middlewares := mid.ServerMiddlewares(midOpt.TracerProvider, mids...)
	opts := []grpc.ServerOption{
		grpc.Middleware(middlewares...),
	}
	if grpcOpt.Network != "" {
		opts = append(opts, grpc.Network(grpcOpt.Network))
	}
	if grpcOpt.Addr != "" {
		opts = append(opts, grpc.Address(grpcOpt.Addr))
	}
	if grpcOpt.Timeout != nil {
		opts = append(opts, grpc.Timeout(grpcOpt.Timeout.AsDuration()))
	}
	return opts
}

// grpc客户端选项
func NewGrpcClientOpt(grpcOpt *TransporterOpt, midOpt *MiddlewareOpt) []grpc.ClientOption {
	if grpcOpt.ServiceName == "" {
		panic("grpcOpt.ServiceName is empty")
	}
	if grpcOpt.Discovery == nil {
		panic("grpcOpt.Discovery is nil")
	}
	if grpcOpt.Timeout == nil {
		panic("grpcOpt.Timeout is nil")
	}
	mids := []mid.MiddlewareOption{}
	if midOpt.TracerProvider == nil {
		panic("NewGrpcServerOpt TracerProvider is nil")
	}

	if midOpt.ServiceKey != "" {
		mids = append(mids, mid.WithAuthMiddleware(mid.Auth{
			ServiceKey: midOpt.ServiceKey,
		}))
	}
	middlewares := mid.ClientMiddlewares(midOpt.TracerProvider, mids...)
	opts := []grpc.ClientOption{
		grpc.WithEndpoint(fmt.Sprintf("discovery:///%v.%v", os.Getenv("SERVICE_ENV"), grpcOpt.ServiceName)),
		grpc.WithDiscovery(grpcOpt.Discovery),
		grpc.WithMiddleware(middlewares...),
		grpc.WithTimeout(grpcOpt.Timeout.AsDuration()),
	}
	return opts
}

type HttpServer struct {
	HttpServer *http.Server
}

type GrpcServer struct {
	GrpcServer *grpc.Server
}

type TaskServer struct {
	TaskServer *asynq.ServeMux
}
