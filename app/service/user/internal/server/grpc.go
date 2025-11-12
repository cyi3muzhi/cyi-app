package server

import (
	"github.com/cin/cin-app/app/service/user/internal/conf"
	"github.com/cin/cin-app/app/service/user/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	userSrv "github.com/cin/cin-app/api/service/user/v1"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server,
	userService *service.UserService,
	logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
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
