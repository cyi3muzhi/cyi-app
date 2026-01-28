package server

import (
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/conf"
	"cyi-app/app/cin/interface/internal/service"
	mid "cyi-app/pkg/util/server/middleware"
	ts "cyi-app/pkg/util/server/transporter"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/trace"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server,
	ac *conf.Auth,
	logger log.Logger,
	tp trace.TracerProvider,
	cinInterface *service.CinInterface,
) *grpc.Server {
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
	v1.RegisterCinInterfaceServer(srv, cinInterface)

	return srv
}
