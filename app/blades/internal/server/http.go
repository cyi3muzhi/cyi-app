package server

import (
	"cyi-app/app/blades/internal/service"

	v1 "cyi-app/api/blades/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(s *ts.HttpServer, blades *service.BladesService) *http.Server {
	v1.RegisterBladesHTTPServer(s.HttpServer, blades)
	// s.HttpServer.HandleFunc("/ws/chat", cinInterface.UserChat)
	return s.HttpServer
}
