package middleware

import (
	validate "github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
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

type MiddlewareOptions struct {
	Auth Auth
	CS   bool // 默认服务端 传true为客户端
}

type MiddlewareOption func(*MiddlewareOptions)

type Auth struct {
	ApiKey     string
	ServiceKey string
	WhiteList  []string
}

// 添加认证中间件
func WithAuthMiddleware(auth Auth) MiddlewareOption {
	return func(o *MiddlewareOptions) {
		o.Auth = auth
	}
}

func WithCS(cs bool) MiddlewareOption {
	return func(o *MiddlewareOptions) {
		o.CS = cs
	}
}

// GetDefaultMiddlewares 获取默认中间件列表
// 包含必备的中间件，如recovery和validate
func GetMiddlewares(tp trace.TracerProvider, logger log.Logger, opts ...MiddlewareOption) []middleware.Middleware {
	options := &MiddlewareOptions{
		CS: false,
		Auth: Auth{
			ApiKey:     "",
			ServiceKey: "",
			WhiteList:  []string{},
		},
	}
	for _, opt := range opts {
		opt(options)
	}
	var middlewareList []middleware.Middleware
	if options.CS {
		if options.Auth.ServiceKey != "" {
			middlewareList = append(middlewareList,
				recovery.Recovery(),
				tracing.Client(tracing.WithTracerProvider(tp)),
			)
			middlewareList = append(middlewareList,
				AuthClientMiddleware(options.Auth),
			)
		}
	} else {
		middlewareList = append(middlewareList,
			recovery.Recovery(),
			tracing.Server(
				tracing.WithTracerProvider(tp)),
			logging.Server(logger),
		)
		if options.Auth.ApiKey != "" {
			middlewareList = append(middlewareList, AuthServiceMiddleware(options.Auth))
		}
		middlewareList = append(middlewareList,
			validate.ProtoValidate(),
			ratelimit.Server(),
		)
	}

	return middlewareList
}
