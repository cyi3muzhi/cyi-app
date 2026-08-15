// Package httproute 提供 Kratos HTTP 服务器手动注册路由的通用工具。
// proto 无法表达 SSE 流式接口与 multipart 上传接口，需要手动 HandleFunc 注册，
// 且 HandleFunc 无法直接套用 Kratos 中间件。本包统一解决这两类问题，
// 供 admin、interface(BFF) 等服务复用，避免重复实现。
package httproute

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"strings"

	mid "cyi-app/pkg/kratos/server/middleware"
	kratosService "cyi-app/pkg/kratos/service"

	"github.com/go-kratos/kratos/v2/middleware"
	jwtmiddleware "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// WithMiddlewares 在 HandleFunc 前执行中间件链，并将处理后的 context（含 JWT claims）传给 handler
func WithMiddlewares(mws []middleware.Middleware, next func(nethttp.ResponseWriter, *nethttp.Request)) func(nethttp.ResponseWriter, *nethttp.Request) {
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

// JWTAuthMiddleware 创建 JWT 中间件（secret 为签名密钥，由各服务 conf.Auth.ApiKey 提供）
// 与 proto 注册接口使用的 mid.AuthServiceMiddleware 保持同一套 claims（UserClaims）
func JWTAuthMiddleware(secret []byte) middleware.Middleware {
	if len(secret) == 0 {
		return func(next middleware.Handler) middleware.Handler {
			return next
		}
	}
	return jwtmiddleware.Server(func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	}, jwtmiddleware.WithSigningMethod(jwtv5.SigningMethodHS256), jwtmiddleware.WithClaims(func() jwtv5.Claims {
		return &mid.UserClaims{}
	}))
}

// ChatStreamService 流式聊天服务接口，由各服务的 service 层实现。
// 服务层只负责业务逻辑（拼接 prompt、调用大模型、输出 chunk），HTTP/SSE 细节由注册器处理。
type ChatStreamService interface {
	// ChatSSE 处理流式聊天
	// message: 用户输入的消息
	// userID: 用户 ID（已从 JWT 解析）
	// emit: 向客户端写入一条 chunk 事件数据并 flush
	ChatSSE(ctx context.Context, message string, userID int64, emit func([]byte)) error
}

// RegisterChatStream 注册 SSE 流式聊天路由（proto 无法定义流式接口，手动注册）。
// path 示例："/admin/v1/agent/chat/stream"。
// 请求体：{"message": "..."}；事件流：start -> chunk* -> done，出错时发送 error 事件。
func RegisterChatStream(s *kratosHttp.Server, path string, mws []middleware.Middleware, svc ChatStreamService) {
	s.HandleFunc(path, WithMiddlewares(mws, func(w nethttp.ResponseWriter, r *nethttp.Request) {
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

		userID, err := kratosService.GetUserID(r.Context())
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

		// 处理流式聊天
		if err := svc.ChatSSE(r.Context(), req.Message, userID, func(data []byte) {
			// 只过滤真正空的 chunk（nil 或空字节），保留换行符等空白字符
			// 因为 Markdown 格式依赖空行（\n\n）来分隔标题、表格等元素
			if len(data) == 0 {
				return
			}
			fmt.Fprintf(w, "event: chunk\ndata: %s\n\n", data)
			flusher.Flush()
		}); err != nil {
			// 不向客户端回显内部错误细节，避免泄露内部信息
			fmt.Fprintf(w, "event: error\ndata: {\"message\":\"ai service unavailable\"}\n\n")
			flusher.Flush()
		}

		// 发送完成事件
		fmt.Fprintf(w, "event: done\ndata: {}\n\n")
		flusher.Flush()
	}))
}

// KnowledgeUploadService 知识库上传服务接口，由各服务的 service 层实现。
type KnowledgeUploadService interface {
	// UploadKnowledgeFile 上传知识库文件
	// 返回文档 ID 与分块数量
	UploadKnowledgeFile(ctx context.Context, filename, source, title, contentType string, content []byte, tags map[string]string) (string, int, error)
}

// RegisterKnowledgeUpload 注册知识库文件上传路由（proto 不支持 multipart，手动注册）。
// path 示例："/admin/v1/agent/knowledge/upload"。
// multipart 字段：file（文件）、source、title、content_type、tags（JSON 格式 map）。
func RegisterKnowledgeUpload(s *kratosHttp.Server, path string, mws []middleware.Middleware, svc KnowledgeUploadService) {
	s.HandleFunc(path, WithMiddlewares(mws, func(w nethttp.ResponseWriter, r *nethttp.Request) {
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
			contentType = DetectContentType(header.Filename)
		}

		// 解析可选的 tags 参数（JSON 格式）
		tags := make(map[string]string)
		if tagsStr := r.FormValue("tags"); tagsStr != "" {
			json.Unmarshal([]byte(tagsStr), &tags)
		}

		// 调用 service 层上传
		docID, chunkCount, err := svc.UploadKnowledgeFile(r.Context(), header.Filename, source, title, contentType, content, tags)
		if err != nil {
			// 不向客户端回显内部错误细节，避免泄露内部信息
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(nethttp.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    500,
				"message": "文件上传失败，请稍后重试",
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

// DetectContentType 根据文件扩展名推断内容类型
func DetectContentType(filename string) string {
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
