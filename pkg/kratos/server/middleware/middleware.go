package middleware

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"go.opentelemetry.io/otel/trace"
)

// 获取默认中间件列表
func Middlewares(tp trace.TracerProvider, middlewareType MiddlewareType, opts ...MiddlewareOption) []middleware.Middleware {
	options := &MiddlewareOptions{
		Auth: Auth{
			ApiKey:     "",
			ServiceKey: "",
			WhiteList:  []string{},
		},
		logger: nil,
		Meter: &Meter{
			Requests: nil,
			Seconds:  nil,
		},
	}
	for _, opt := range opts {
		opt(options)
	}
	var middlewareList []middleware.Middleware
	if middlewareType {
		middlewareList = append(middlewareList,
			recovery.Recovery(),
			tracing.Client(tracing.WithTracerProvider(tp)),
		)
		if options.Meter.Requests != nil || options.Meter.Seconds != nil {
			middlewareList = append(middlewareList, metrics.Client(
				metrics.WithSeconds(options.Meter.Seconds),
				metrics.WithRequests(options.Meter.Requests),
			))
		}

		if options.Auth.ServiceKey != "" {
			middlewareList = append(middlewareList,
				AuthClientMiddleware(options.Auth),
			)
		}
	} else {
		middlewareList = append(middlewareList,
			recovery.Recovery(),
			tracing.Server(
				tracing.WithTracerProvider(tp),
			),
			TraceIDHeader(),
		)
		if options.logger != nil {
			middlewareList = append(middlewareList, logging.Server(options.logger))
		}
		if options.Meter.Requests != nil || options.Meter.Seconds != nil {
			middlewareList = append(middlewareList, metrics.Server(
				metrics.WithSeconds(options.Meter.Seconds),
				metrics.WithRequests(options.Meter.Requests),
			))
		}

		if options.Auth.ApiKey != "" {
			middlewareList = append(middlewareList, AuthServiceMiddleware(options.Auth))
		}
		if options.Enforcer != nil {
			middlewareList = append(middlewareList, PermissionMiddleware(options.Enforcer))
		}
		middlewareList = append(middlewareList,
			Validate(),
			ratelimit.Server(),
		)
	}

	return middlewareList
}

// 客户端中间件列表
func ClientMiddlewares(tp trace.TracerProvider, opts ...MiddlewareOption) []middleware.Middleware {
	return Middlewares(tp, Client, opts...)
}

// 服务端中间件列表
func ServerMiddlewares(tp trace.TracerProvider, opts ...MiddlewareOption) []middleware.Middleware {
	return Middlewares(tp, Server, opts...)
}
