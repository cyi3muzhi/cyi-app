package middleware

import (
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/metric"
)

const (
	Server = false
	Client = true
)

type Auth struct {
	ApiKey         string
	ServiceKey     string
	WhiteList      []string
	TokenValidator TokenValidator
}

type Meter struct {
	Requests metric.Int64Counter
	Seconds  metric.Float64Histogram
}

type MiddlewareType bool

type MiddlewareOptions struct {
	Auth     Auth
	Enforcer EnforcerProvider
	logger   log.Logger
	Meter    *Meter
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

// 添加指标中间件
func WithMetricsMiddleware(meter*Meter) MiddlewareOption {
	return func(o *MiddlewareOptions) {
		o.Meter = meter
	}
}

// 添加 Casbin 权限中间件
func WithCasbinMiddleware(enforcer EnforcerProvider) MiddlewareOption {
	return func(o *MiddlewareOptions) {
		o.Enforcer = enforcer
	}
}
