package server

import (
	nethttp "net/http"

	"cyi-app/app/cin/interface/internal/conf"
	"cyi-app/app/cin/interface/internal/service"
	httproute "cyi-app/pkg/kratos/server/httproute"

	"github.com/go-kratos/kratos/v2/middleware"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
)

// registerStreamRoutes 注册流式路由（SSE + WebSocket），手工包裹中间件
func registerStreamRoutes(s *kratosHttp.Server, cinInterface *service.CinInterface, ac *conf.Auth) {
	authMw := []middleware.Middleware{httproute.JWTAuthMiddleware([]byte(ac.GetApiKey()))}

	s.HandleFunc("/ai/v1/chat/stream", httproute.WithMiddlewares(authMw, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Content-Type", "text/event-stream")
				nethttp.Error(w, "internal server error", nethttp.StatusInternalServerError)
			}
		}()
		cinInterface.ChatSSE(w, r)
	}))

	s.HandleFunc("/ws/chat", httproute.WithMiddlewares([]middleware.Middleware{}, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		defer func() {
			if err := recover(); err != nil {
				nethttp.Error(w, "internal server error", nethttp.StatusInternalServerError)
			}
		}()
		cinInterface.ChatWS(w, r)
	}))

	// IM 客服 WebSocket（query token 鉴权在 ws.Serve 内部处理，不走 HTTP 中间件）
	s.HandleFunc("/ws/chat/v2", httproute.WithMiddlewares([]middleware.Middleware{}, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		defer func() {
			if err := recover(); err != nil {
				nethttp.Error(w, "internal server error", nethttp.StatusInternalServerError)
			}
		}()
		cinInterface.IMChatWS(w, r)
	}))
}
