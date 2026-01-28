package server

import (
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/conf"
	"cyi-app/app/cin/interface/internal/service"

	cmid "cyi-app/pkg/util/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	auth *conf.Auth,
	authWhiteList *conf.AuthWhiteList,
	transactionList *conf.TransactionList,
	cinInterface *service.CinInterface,
	tx *gorm.DB,
) *http.Server {
	// 合并默认中间件和自定义中间件
	middlewares := cmid.GetMiddlewares(tp, logger,
		cmid.WithAuthMiddleware(cmid.Auth{
			ApiKey:    auth.ApiKey,
			WhiteList: authWhiteList.WhiteList,
		}),
	)
	// 事务管理中间件
	// allMiddlewares = append(allMiddlewares, cmid.TransactionMiddleware(tx, transactionList.Transactions))

	var opts = cmid.NewHttpServerOpt(middlewares...)
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
