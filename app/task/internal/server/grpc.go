package server

import (
	"cyi-app/app/task/internal/service"

	v1 "cyi-app/api/task/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewServer(s *ts.GrpcServer, taskService *service.TaskService) *grpc.Server {
	v1.RegisterTaskServer(s.GrpcServer, taskService)
	return s.GrpcServer
}
