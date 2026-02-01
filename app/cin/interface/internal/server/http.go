package server

import (
	"cyi-app/app/cin/interface/internal/service"

	v1 "cyi-app/api/cin/interface/v1"

	"github.com/go-kratos/kratos/v2/transport/http"
	ts "cyi-app/pkg/kratos/server/transporter"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(s *ts.HttpServer, cinInterface *service.CinInterface) *http.Server {
	v1.RegisterCinInterfaceHTTPServer(s.HttpServer, cinInterface)
	return s.HttpServer
}
