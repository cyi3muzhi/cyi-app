package server

import (
	"cyi-app/app/service/user/internal/conf"
	"cyi-app/app/service/user/internal/service"

	userSrv "cyi-app/api/service/user/v1"

	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/trace"
)

func GetDefaultMiddlewares2(tp trace.TracerProvider, logger log.Logger, ac *conf.Auth) []middleware.Middleware {
	return []middleware.Middleware{
		recovery.Recovery(),
		tracing.Server(
			tracing.WithTracerProvider(tp)),
		metrics.Server(),
		logging.Server(logger),
		jwt.Server(func(token *jwt2.Token) (interface{}, error) {
			return []byte(ac.ApiKey), nil
		}, jwt.WithSigningMethod(jwt2.SigningMethodHS256)),
		validate.ProtoValidate(),
	}
}

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server,
	userService *service.UserService,
	tp trace.TracerProvider,
	ac *conf.Auth,
	logger log.Logger) *grpc.Server {
	allMiddlewares := GetDefaultMiddlewares2(tp, logger, ac)
	var opts = []grpc.ServerOption{
		grpc.Middleware(allMiddlewares...),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	userSrv.RegisterUserServer(srv, userService)

	return srv
}
