package server

import (
	"cyi-app/app/service/user/internal/service"

	v1 "cyi-app/api/service/user/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewServer(s *ts.GrpcServer, userService *service.UserService, adminUserService *service.AdminUserService) *grpc.Server {
	v1.RegisterUserServer(s.GrpcServer, userService)
	v1.RegisterAdminUserServer(s.GrpcServer, adminUserService)
	return s.GrpcServer
}
