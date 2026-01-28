package middleware

import (
	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"go.opentelemetry.io/otel/trace"
)

//  获取默认中间件列表
func Middlewares(tp trace.TracerProvider, middlewareType MiddlewareType, opts ...MiddlewareOption) []middleware.Middleware {
	options := &MiddlewareOptions{
		Auth: Auth{
			ApiKey:     "",
			ServiceKey: "",
			WhiteList:  []string{},
		},
		logger: nil,
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
		if options.Auth.ServiceKey != "" {
			middlewareList = append(middlewareList,
				AuthClientMiddleware(options.Auth),
			)
		}
	} else {
		middlewareList = append(middlewareList,
			recovery.Recovery(),
			tracing.Server(
				tracing.WithTracerProvider(tp)),
		)
		if options.logger != nil {
			middlewareList = append(middlewareList, logging.Server(options.logger))
		}
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

func ClientMiddlewares(tp trace.TracerProvider, opts ...MiddlewareOption) []middleware.Middleware {
	return Middlewares(tp, Client, opts...)
}

func ServerMiddlewares(tp trace.TracerProvider, opts ...MiddlewareOption) []middleware.Middleware {
	return Middlewares(tp, Server, opts...)
}