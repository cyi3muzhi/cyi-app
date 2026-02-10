package server

import (
	"cyi-app/app/task/internal/service"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(s *ts.HttpServer, taskService *service.TaskService) *http.Server {
	// v1.RegisterHTTPTaskServer(s.HttpServer, taskService)
	return s.HttpServer
}
