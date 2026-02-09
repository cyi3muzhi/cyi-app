package server

import (
	"cyi-app/app/cin/interface/internal/service"

	v1 "cyi-app/api/cin/interface/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(s *ts.HttpServer, cinInterface *service.CinInterface) *http.Server {
	v1.RegisterCinInterfaceHTTPServer(s.HttpServer, cinInterface)
	s.HttpServer.HandleFunc("/ws/chat", cinInterface.UserChat)
	return s.HttpServer
}
