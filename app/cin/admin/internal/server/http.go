package server

import (
	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/conf"
	"cyi-app/app/cin/admin/internal/service"
	httproute "cyi-app/pkg/kratos/server/httproute"
	nethttp "net/http"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/middleware"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server. 注册管理后台的多个接口服务
func NewHTTPServer(s *ts.HttpServer, adminService *service.AdminService, logsService *service.LogsService, ac *conf.Auth) *kratosHttp.Server {
	adminv1.RegisterDashboardHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminProductHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminOrderHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminUserHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterSettingsHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminAuthHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminRoleHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterStaffHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminCategoryHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterUploadHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminCouponHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterBannerHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterAdminMessageHTTPServer(s.HttpServer, adminService)
	adminv1.RegisterChatHTTPServer(s.HttpServer, adminService)
	// 注册 Agent proto 接口（替代 registerAgentRoutes 中除 SSE 外的所有路由）
	adminv1.RegisterAgentHTTPServer(s.HttpServer, adminService)
	// 注册操作日志查询接口（独立服务）
	adminv1.RegisterAdminLogsHTTPServer(s.HttpServer, logsService)

	// 注册 SSE 流式聊天路由（proto 不支持流式，手动注册）
	authMw := []middleware.Middleware{httproute.JWTAuthMiddleware([]byte(ac.GetApiKey()))}
	httproute.RegisterChatStream(s.HttpServer, "/admin/v1/agent/chat/stream", authMw, adminService)
	// 注册知识库文件上传路由（proto 不支持 multipart，手动注册）
	httproute.RegisterKnowledgeUpload(s.HttpServer, "/admin/v1/agent/knowledge/upload", authMw, adminService)

	// 注册客服 IM WebSocket（query token 鉴权在 ws.Serve 内部处理，不走 HTTP 中间件）
	s.HttpServer.HandleFunc("/admin/v1/chat/ws", httproute.WithMiddlewares([]middleware.Middleware{}, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		defer func() {
			if err := recover(); err != nil {
				nethttp.Error(w, "internal server error", nethttp.StatusInternalServerError)
			}
		}()
		adminService.IMAgentChatWS(w, r)
	}))
	return s.HttpServer
}
