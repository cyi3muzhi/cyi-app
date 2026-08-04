package middleware

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"time"

	"cyi-app/pkg/kratos/event"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	jwtmiddleware "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// OperationLogRecorder 操作日志落库接口（由业务侧实现，如发布到 NATS）
type OperationLogRecorder interface {
	Publish(evt *event.AdminOperationLog)
}

// OperationLog 操作日志中间件
// 对非 GET 写操作自动记录，通过 recorder 异步落库（失败不影响主流程）
// 需在认证中间件之后执行，以从 JWT claims 中获取操作人信息
func OperationLog(recorder OperationLogRecorder, logger log.Logger) middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()

			// 只记录已登录管理员的写操作（GET 查询不记录）
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return next(ctx, req)
			}
			// HTTP 方法/真实路径/IP 从请求对象获取（RequestHeader 不含 method；Operation() 是 gRPC 网关方法名而非 URL）
			method := "GET"
			var reqPath, remoteAddr string
			if htr, ok := tr.(khttp.Transporter); ok && htr.Request() != nil {
				method = htr.Request().Method
				reqPath = htr.Request().URL.Path
				remoteAddr = htr.Request().RemoteAddr
			}
			if method == "GET" {
				return next(ctx, req)
			}

			// 提取操作人（JWT claims）
			claims, ok := jwtmiddleware.FromContext(ctx)
			if !ok {
				return next(ctx, req)
			}
			userClaims, ok := claims.(*UserClaims)
			if !ok || userClaims == nil {
				return next(ctx, req)
			}

			path := reqPath
			if path == "" {
				path = tr.Operation()
			}

			// 执行真实业务
			resp, err := next(ctx, req)
			durationMs := time.Since(start).Milliseconds()

			// 记录日志（异步发布，失败不影响主流程）
			evt := &event.AdminOperationLog{
				AdminUserID:  userClaims.ID,
				OperatorName: firstNonEmpty(userClaims.Name, userClaims.Username),
				RoleName:     userClaims.Role,
				Module:       detectModule(path),
				Action:       detectAction(method),
				Content:      buildContent(method, path),
				Method:       method,
				Path:         path,
				IP:           normalizeIP(tr.RequestHeader().Get("X-Forwarded-For"), tr.RequestHeader().Get("X-Real-IP"), remoteAddr),
				UserAgent:    tr.RequestHeader().Get("User-Agent"),
				RequestBody:  truncateBody(req),
				DurationMs:   durationMs,
				CreatedAt:    time.Now().Unix(),
			}
			if err != nil {
				evt.Result = 0
			} else {
				evt.Result = 1
			}
			if recorder != nil {
				go recorder.Publish(evt)
			}

			return resp, err
		}
	}
}

// detectModule 根据请求路径识别操作模块
func detectModule(path string) string {
	switch {
	case strings.Contains(path, "/admin/v1/products"):
		return "product"
	case strings.Contains(path, "/admin/v1/orders"):
		return "order"
	case strings.Contains(path, "/admin/v1/users"):
		return "user"
	case strings.Contains(path, "/admin/v1/coupons"):
		return "coupon"
	case strings.Contains(path, "/admin/v1/categories"):
		return "category"
	case strings.Contains(path, "/admin/v1/banners"):
		return "banner"
	case strings.Contains(path, "/admin/v1/staff"):
		return "staff"
	case strings.Contains(path, "/admin/v1/roles"):
		return "role"
	case strings.Contains(path, "/admin/v1/settings"):
		return "setting"
	default:
		return "other"
	}
}

// detectAction 根据 HTTP 方法识别动作
func detectAction(method string) string {
	switch method {
	case "POST":
		return "新增"
	case "PUT", "PATCH":
		return "编辑"
	case "DELETE":
		return "删除"
	default:
		return method
	}
}

// buildContent 生成操作内容描述
func buildContent(method, path string) string {
	return detectAction(method) + " " + path
}

// truncateBody 序列化请求体并截断（限制长度，避免日志过大）
func truncateBody(req interface{}) string {
	if req == nil {
		return ""
	}
	data, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	text := string(data)
	if len(text) > 500 {
		return text[:500] + "..."
	}
	return text
}

// firstNonEmpty 返回第一个非空字符串
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// normalizeIP 规范化来源 IP：优先代理头，其次 RemoteAddr（去掉端口）
func normalizeIP(forwardedFor, realIP, remoteAddr string) string {
	if ip := firstNonEmpty(forwardedFor, realIP); ip != "" {
		// X-Forwarded-For 可能为逗号分隔的多个 IP，取第一个
		if idx := strings.IndexByte(ip, ','); idx >= 0 {
			ip = strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}
	if remoteAddr == "" {
		return ""
	}
	// RemoteAddr 形如 "127.0.0.1:8080" 或 "[::1]:57243"，去掉端口
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}
