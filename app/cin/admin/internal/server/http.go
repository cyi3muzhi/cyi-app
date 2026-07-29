package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"strings"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/conf"
	"cyi-app/app/cin/admin/internal/service"
	mid "cyi-app/pkg/kratos/server/middleware"
	kratosService "cyi-app/pkg/kratos/service"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/middleware"
	jwtmiddleware "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// NewHTTPServer new an HTTP server. 注册管理后台的多个接口服务
func NewHTTPServer(s *ts.HttpServer, adminService *service.AdminService, ac *conf.Auth) *kratosHttp.Server {
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
	// 注册 Agent proto 接口（替代 registerAgentRoutes 中除 SSE 外的所有路由）
	adminv1.RegisterAgentHTTPServer(s.HttpServer, adminService)
	// 注册 SSE 流式聊天路由（proto 不支持流式，手动注册）
	registerAgentSSEServer(s.HttpServer, adminService, ac)
	// 注册知识库文件上传路由（proto 不支持 multipart，手动注册）
	registerKnowledgeUploadServer(s.HttpServer, adminService, ac)
	return s.HttpServer
}

// registerAgentSSEServer 注册 SSE 流式聊天路由（proto 无法定义流式接口，手动注册）
func registerAgentSSEServer(s *kratosHttp.Server, adminService *service.AdminService, ac *conf.Auth) {
	authMw := []middleware.Middleware{newJWTAuthMiddleware(ac)}

	// POST /admin/v1/agent/chat/stream SSE 流式聊天
	s.HandleFunc("/admin/v1/agent/chat/stream", withMiddlewares(authMw, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method != nethttp.MethodPost {
			nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			nethttp.Error(w, "读取请求体失败", nethttp.StatusBadRequest)
			return
		}
		var req struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			nethttp.Error(w, "无效的 JSON 格式", nethttp.StatusBadRequest)
			return
		}
		if req.Message == "" {
			nethttp.Error(w, "消息不能为空", nethttp.StatusBadRequest)
			return
		}

		adminID, err := kratosService.GetUserID(r.Context())
		if err != nil {
			nethttp.Error(w, "未登录", nethttp.StatusUnauthorized)
			return
		}

		// 设置 SSE 响应头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(nethttp.Flusher)
		if !ok {
			nethttp.Error(w, "Streaming not supported", nethttp.StatusInternalServerError)
			return
		}

		// 发送开始事件
		fmt.Fprintf(w, "event: start\ndata: {}\n\n")
		flusher.Flush()

		// 使用 ChatSSEHandler 处理流式聊天
		_, err = adminService.ChatSSEHandler(r.Context(), req.Message, adminID, func(data []byte) {
			// 过滤空白 chunk
			var chunkData struct {
				Chunk string `json:"chunk"`
			}
			if json.Unmarshal(data, &chunkData) == nil && strings.TrimSpace(chunkData.Chunk) == "" {
				return
			}
			fmt.Fprintf(w, "event: chunk\ndata: %s\n\n", data)
			flusher.Flush()
		})
		if err != nil {
			// 向客户端暴露原始错误以便排查
			fmt.Fprintf(w, "event: error\ndata: {\"message\":\"%s\"}\n\n", err.Error())
			flusher.Flush()
		}

		// 发送完成事件
		fmt.Fprintf(w, "event: done\ndata: {}\n\n")
		flusher.Flush()
	}))
}

// withMiddlewares 在 HandleFunc 前执行中间件链，并将处理后的 context 传给 handler
func withMiddlewares(mws []middleware.Middleware, next func(nethttp.ResponseWriter, *nethttp.Request)) func(nethttp.ResponseWriter, *nethttp.Request) {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var claimsCtx context.Context
		handler := middleware.Chain(mws...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			claimsCtx = ctx
			return nil, nil
		})
		if _, err := handler(r.Context(), nil); err != nil {
			nethttp.Error(w, err.Error(), nethttp.StatusUnauthorized)
			return
		}
		// 使用中间件处理后的 context（含 JWT claims），替换 request 中的 context
		next(w, r.WithContext(claimsCtx))
	}
}

// newJWTAuthMiddleware 创建 JWT 中间件（复用 BFF 的 auth.ApiKey）
func newJWTAuthMiddleware(ac *conf.Auth) middleware.Middleware {
	if ac == nil || ac.ApiKey == "" {
		return func(next middleware.Handler) middleware.Handler {
			return next
		}
	}
	return jwtmiddleware.Server(func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(ac.ApiKey), nil
	}, jwtmiddleware.WithSigningMethod(jwtv5.SigningMethodHS256), jwtmiddleware.WithClaims(func() jwtv5.Claims {
		return &mid.UserClaims{}
	}))
}

// registerKnowledgeUploadServer 注册知识库文件上传路由（multipart form 解析，手动注册）
func registerKnowledgeUploadServer(s *kratosHttp.Server, adminService *service.AdminService, ac *conf.Auth) {
	authMw := []middleware.Middleware{newJWTAuthMiddleware(ac)}

	// POST /admin/v1/agent/knowledge/upload 知识库文件上传
	s.HandleFunc("/admin/v1/agent/knowledge/upload", withMiddlewares(authMw, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method != nethttp.MethodPost {
			nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
			return
		}

		// 限制上传大小（50MB）
		r.Body = nethttp.MaxBytesReader(w, r.Body, 50<<20)

		// 解析 multipart form
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			nethttp.Error(w, "文件大小超出限制或解析失败", nethttp.StatusBadRequest)
			return
		}

		// 获取上传文件
		file, header, err := r.FormFile("file")
		if err != nil {
			nethttp.Error(w, "未找到上传文件", nethttp.StatusBadRequest)
			return
		}
		defer file.Close()

		// 读取文件内容
		content, err := io.ReadAll(file)
		if err != nil {
			nethttp.Error(w, "读取文件内容失败", nethttp.StatusInternalServerError)
			return
		}

		// 获取表单参数
		source := r.FormValue("source")
		if source == "" {
			source = "admin"
		}
		title := r.FormValue("title")
		if title == "" {
			title = header.Filename
		}
		contentType := r.FormValue("content_type")
		if contentType == "" {
			contentType = detectContentType(header.Filename)
		}

		// 解析可选的 tags 参数（JSON 格式）
		tags := make(map[string]string)
		if tagsStr := r.FormValue("tags"); tagsStr != "" {
			json.Unmarshal([]byte(tagsStr), &tags)
		}

		// 调用 service 层上传
		docID, chunkCount, err := adminService.UploadKnowledgeFile(r.Context(), header.Filename, source, title, contentType, content, tags)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(nethttp.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"doc_id":      docID,
				"chunk_count": chunkCount,
				"title":       title,
			},
		})
	}))
}

// detectContentType 根据文件扩展名推断内容类型
func detectContentType(filename string) string {
	ext := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(ext, ".pdf"):
		return "pdf"
	case strings.HasSuffix(ext, ".docx") || strings.HasSuffix(ext, ".doc"):
		return "docx"
	case strings.HasSuffix(ext, ".txt"):
		return "txt"
	case strings.HasSuffix(ext, ".md") || strings.HasSuffix(ext, ".markdown"):
		return "markdown"
	case strings.HasSuffix(ext, ".json"):
		return "json"
	case strings.HasSuffix(ext, ".html"):
		return "html"
	default:
		return "unknown"
	}
}
