package server

import (
	"cyi-app/app/cin/interface/internal/conf"
	"cyi-app/app/cin/interface/internal/service"

	v1 "cyi-app/api/cin/interface/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(s *ts.HttpServer, cinInterface *service.CinInterface, ac *conf.Auth) *kratosHttp.Server {
	v1.RegisterCinInterfaceHTTPServer(s.HttpServer, cinInterface)
	registerStreamRoutes(s.HttpServer, cinInterface, ac)
	return s.HttpServer
}
