package server

import (
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/conf"
	"cyi-app/app/cin/interface/internal/service"
	mid "cyi-app/pkg/util/server/middleware"
	ts "cyi-app/pkg/util/server/transporter"

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
	midList := mid.ServerMiddlewares(tp,
		mid.WithAuthMiddleware(mid.Auth{
			ApiKey:    auth.ApiKey,
			WhiteList: authWhiteList.WhiteList,
		}),
		mid.WithLoggerMiddleware(logger),
	)

	var opts = ts.NewHttpServerOpt(
		ts.TransporterOpt{
			Network: c.Http.Network,
			Addr:    c.Http.Addr,
			Timeout: c.Http.Timeout,
		},
		midList...)
	srv := http.NewServer(opts...)
	v1.RegisterCinInterfaceHTTPServer(srv, cinInterface)
	return srv
}

func RegisterHTTPServer(
	srv *http.Server,
	cinInterface *service.CinInterface,
) {
}
