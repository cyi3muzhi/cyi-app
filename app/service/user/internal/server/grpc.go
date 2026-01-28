package server

import (
	"cyi-app/app/service/user/internal/conf"
	"cyi-app/app/service/user/internal/service"

	v1 "cyi-app/api/service/user/v1"

	mid "cyi-app/pkg/util/server/middleware"
	ts "cyi-app/pkg/util/server/transporter"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/trace"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server,
	userService *service.UserService,
	tp trace.TracerProvider,
	ac *conf.Auth,
	logger log.Logger) *grpc.Server {
	midList := mid.ServerMiddlewares(tp,
		mid.WithAuthMiddleware(mid.Auth{
			ApiKey: ac.ApiKey,
		}),
		mid.WithLoggerMiddleware(logger),
	)
	var opts = ts.NewGrpcServerOpt(ts.TransporterOpt{
		Network: c.Grpc.Network,
		Addr:    c.Grpc.Addr,
		Timeout: c.Grpc.Timeout,
	}, midList...)

	srv := grpc.NewServer(opts...)
	v1.RegisterUserServer(srv, userService)

	return srv
}
