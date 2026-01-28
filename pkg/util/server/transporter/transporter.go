package transporter

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"google.golang.org/protobuf/types/known/durationpb"
)

type TransporterOpt struct {
	Network string
	Addr    string
	Timeout *durationpb.Duration
}
//  http服务器选项
func NewHttpServerOpt(httpOpt TransporterOpt, middlewares ...middleware.Middleware) []http.ServerOption {
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
func NewGrpcServerOpt(grpcOpt TransporterOpt, middlewares ...middleware.Middleware) []grpc.ServerOption {
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

//  grpc客户端选项
func NewGrpcClientOpt(grpcOpt TransporterOpt, middlewares ...middleware.Middleware) []grpc.ClientOption {
	opts := []grpc.ClientOption{
		grpc.WithMiddleware(middlewares...),
		grpc.WithTimeout(grpcOpt.Timeout.AsDuration()),
	}
	return opts
}
