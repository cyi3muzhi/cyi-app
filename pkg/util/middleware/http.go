package middleware

import (
	validate "github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"go.opentelemetry.io/otel/trace"
)

func NewHttpServerOpt(middlewares ...middleware.Middleware) []http.ServerOption {
	return []http.ServerOption{
		http.Middleware(middlewares...),
		http.Filter(handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}),
		)),
	}
}

// GetDefaultMiddlewares 获取默认中间件列表
// 包含必备的中间件，如recovery和validate
func GetDefaultMiddlewares(tp trace.TracerProvider, logger log.Logger, apiKey string, authWhiteList []string) []middleware.Middleware {
	return []middleware.Middleware{
		recovery.Recovery(),
		tracing.Server(
			tracing.WithTracerProvider(tp)),
		logging.Server(logger),
		selector.Server(
			jwt.Server(func(token *jwt2.Token) (interface{}, error) {
				return []byte(apiKey), nil
			}, jwt.WithSigningMethod(jwt2.SigningMethodHS256), jwt.WithClaims(func() jwt2.Claims {
				return &jwt2.MapClaims{}
			})),
		).
			Match(NewWhiteListMatcher(authWhiteList)).
			Build(),
		validate.ProtoValidate(),
		ratelimit.Server(),
	}
}
