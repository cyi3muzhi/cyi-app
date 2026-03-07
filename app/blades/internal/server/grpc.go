package server

import (
	v1 "cyi-app/api/blades/v1"
	"cyi-app/app/blades/internal/service"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	ts "cyi-app/pkg/kratos/server/transporter"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(s *ts.GrpcServer, blades *service.BladesService) *grpc.Server {
	v1.RegisterBladesServer(s.GrpcServer, blades)
	return s.GrpcServer
}
