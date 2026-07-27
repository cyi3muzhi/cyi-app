package server

import (
	"cyi-app/app/cin/interface/internal/conf"
	"cyi-app/app/cin/interface/internal/service"

	v1 "cyi-app/api/cin/interface/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server. 注册拆分后的多个接口服务
func NewHTTPServer(s *ts.HttpServer, cinInterface *service.CinInterface, ac *conf.Auth) *kratosHttp.Server {
	v1.RegisterAuthHTTPServer(s.HttpServer, cinInterface)
	v1.RegisterUserHTTPServer(s.HttpServer, cinInterface)
	v1.RegisterProductHTTPServer(s.HttpServer, cinInterface)
	v1.RegisterOrderHTTPServer(s.HttpServer, cinInterface)
	v1.RegisterCartHTTPServer(s.HttpServer, cinInterface)
	v1.RegisterStockHTTPServer(s.HttpServer, cinInterface)
	v1.RegisterBladesHTTPServer(s.HttpServer, cinInterface)
	v1.RegisterUserCouponHTTPServer(s.HttpServer, cinInterface)
	v1.RegisterSettingsHTTPServer(s.HttpServer, cinInterface)
	v1.RegisterBannerHTTPServer(s.HttpServer, cinInterface)
	registerStreamRoutes(s.HttpServer, cinInterface, ac)
	return s.HttpServer
}
