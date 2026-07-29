package server

import (
	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer 创建 HTTP 服务器
// Blades 服务仅通过 gRPC 提供服务，HTTP 服务器仅用于 /metrics 端点
func NewHTTPServer(s *ts.HttpServer) *http.Server {
	// 路由由 protoc-gen-go-http 生成时自动注册，当前仅保留 /metrics（已在 server.go 中注册）
	return s.HttpServer
}
