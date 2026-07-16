package server

import (
	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/service"

	ts "cyi-app/pkg/kratos/server/transporter"

	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server. 注册管理后台的多个接口服务
func NewHTTPServer(s *ts.HttpServer, adminService *service.AdminService) *kratosHttp.Server {
	adminv1.RegisterDashboardHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminProductHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminOrderHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminUserHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterSettingsHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminAuthHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminRoleHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterStaffHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminCategoryHTTPServer(s.HttpServer, adminService)
	return s.HttpServer
}
