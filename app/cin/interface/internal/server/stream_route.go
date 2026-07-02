package server

import (
	"context"
	"cyi-app/app/cin/interface/internal/conf"
	"cyi-app/app/cin/interface/internal/service"
	"fmt"
	nethttp "net/http"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/middleware"
	jwtmiddleware "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// registerStreamRoutes 注册流式路由（SSE + WebSocket），手工包裹中间件
func registerStreamRoutes(s *kratosHttp.Server, cinInterface *service.CinInterface, ac *conf.Auth) {
	authMw := []middleware.Middleware{newJWTAuthMiddleware(ac)}

	s.HandleFunc("/ai/v1/chat/stream", withMiddlewares(authMw, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Content-Type", "text/event-stream")
				nethttp.Error(w, "internal server error", nethttp.StatusInternalServerError)
			}
		}()
		cinInterface.ChatSSE(w, r)
	}))

	s.HandleFunc("/ws/chat", withMiddlewares([]middleware.Middleware{}, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		defer func() {
			if err := recover(); err != nil {
				nethttp.Error(w, "internal server error", nethttp.StatusInternalServerError)
			}
		}()
		cinInterface.ChatWS(w, r)
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
