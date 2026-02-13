package server

import (
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/service"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	ts "cyi-app/pkg/kratos/server/transporter"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(s *ts.GrpcServer, cinInterface *service.CinInterface) *grpc.Server {
	v1.RegisterCinInterfaceServer(s.GrpcServer, cinInterface)
	return s.GrpcServer
}
