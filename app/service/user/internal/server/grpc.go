package server

import (
	"cyi-app/app/service/user/internal/service"

	v1 "cyi-app/api/service/user/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewServer(s *ts.GrpcServer, userService *service.UserService, messageService *service.MessageService, adminUserService *service.AdminUserService, adminRoleService *service.AdminRoleService, adminSettingsService *service.AdminSettingsService, adminBannerService *service.AdminBannerService, operationLogService *service.OperationLogService) *grpc.Server {
	v1.RegisterUserServer(s.GrpcServer, userService)
	v1.RegisterMessageServer(s.GrpcServer, messageService)
	v1.RegisterAdminUserServer(s.GrpcServer, adminUserService)
	v1.RegisterAdminRoleServer(s.GrpcServer, adminRoleService)
	v1.RegisterAdminSettingsServer(s.GrpcServer, adminSettingsService)
	v1.RegisterAdminBannerServer(s.GrpcServer, adminBannerService)
	v1.RegisterOperationLogServer(s.GrpcServer, operationLogService)
	return s.GrpcServer
}
