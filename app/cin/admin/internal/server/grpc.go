package server

import (
	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/service"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	ts "cyi-app/pkg/kratos/server/transporter"
)

// NewGRPCServer new a gRPC server. 注册管理后台的多个接口服务
func NewGRPCServer(s *ts.GrpcServer, adminService *service.AdminService) *grpc.Server {
	adminv1.RegisterDashboardServer(s.GrpcServer, adminService)
	adminv1.RegisterAdminProductServer(s.GrpcServer, adminService)
	adminv1.RegisterAdminOrderServer(s.GrpcServer, adminService)
	adminv1.RegisterAdminUserServer(s.GrpcServer, adminService)
	adminv1.RegisterSettingsServer(s.GrpcServer, adminService)
	adminv1.RegisterAdminAuthServer(s.GrpcServer, adminService)
	adminv1.RegisterAdminRoleServer(s.GrpcServer, adminService)
	adminv1.RegisterStaffServer(s.GrpcServer, adminService)
	adminv1.RegisterAdminCategoryServer(s.GrpcServer, adminService)
	adminv1.RegisterUploadServer(s.GrpcServer, adminService)
	adminv1.RegisterAdminCouponServer(s.GrpcServer, adminService)
	adminv1.RegisterAdminMessageServer(s.GrpcServer, adminService)
	adminv1.RegisterAgentServer(s.GrpcServer, adminService)
	return s.GrpcServer
}
