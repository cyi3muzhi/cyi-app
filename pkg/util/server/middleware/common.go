package middleware

import "github.com/go-kratos/kratos/v2/log"

const (
	Server = false
	Client = true
)

type Auth struct {
	ApiKey     string
	ServiceKey string
	WhiteList  []string
}

type MiddlewareType bool

type MiddlewareOptions struct {
	Auth   Auth
	logger log.Logger
}

type MiddlewareOption func(*MiddlewareOptions)

// 添加日志中间件
func WithLoggerMiddleware(logger log.Logger) MiddlewareOption {
	return func(o *MiddlewareOptions) {
		o.logger = logger
	}
}

// 添加认证中间件
func WithAuthMiddleware(auth Auth) MiddlewareOption {
	return func(o *MiddlewareOptions) {
		o.Auth = auth
	}
}
