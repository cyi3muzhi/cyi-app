package server

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/conf"
	"cyi-app/app/cin/interface/internal/service"

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

func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList[v1.OperationCinInterfaceLogin] = struct{}{}
	whiteList[v1.OperationCinInterfaceRegister] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// GetDefaultMiddlewares 获取默认中间件列表
// 包含必备的中间件，如recovery和validate
func GetDefaultMiddlewares(tp trace.TracerProvider, logger log.Logger, ac *conf.Auth) []middleware.Middleware {
	return []middleware.Middleware{
		recovery.Recovery(),
		tracing.Server(
			tracing.WithTracerProvider(tp)),
		logging.Server(logger),
		selector.Server(
			jwt.Server(func(token *jwt2.Token) (interface{}, error) {
				return []byte(ac.ApiKey), nil
			}, jwt.WithSigningMethod(jwt2.SigningMethodHS256), jwt.WithClaims(func() jwt2.Claims {
				return &jwt2.MapClaims{}
			})),
		).
			Match(NewWhiteListMatcher()).
			Build(),
		validate.ProtoValidate(),
		ratelimit.Server(),
	}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	auth *conf.Auth,
	cinInterface *service.CinInterface,
) *http.Server {
	// 合并默认中间件和自定义中间件
	allMiddlewares := GetDefaultMiddlewares(tp, logger, auth)
	// allMiddlewares = append(allMiddlewares, customMiddlewares...)

	var opts = []http.ServerOption{
		http.Middleware(allMiddlewares...),
		http.Filter(handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}),
		)),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterCinInterfaceHTTPServer(srv, cinInterface)
	return srv
}
