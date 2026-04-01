package server

import (
	"cyi-app/app/service/order/internal/service"

	v1 "cyi-app/api/service/order/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewServer(s *ts.GrpcServer, orderService *service.OrderService,
) *grpc.Server {
	v1.RegisterOrderServer(s.GrpcServer, orderService)
	return s.GrpcServer
}
