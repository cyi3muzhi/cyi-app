package server

import (
	"cyi-app/app/service/product/internal/service"

	v1 "cyi-app/api/service/product/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewServer(s *ts.GrpcServer, productService *service.ProductService, stockService *service.StockService) *grpc.Server {
	v1.RegisterProductServer(s.GrpcServer, productService)
	v1.RegisterStockServer(s.GrpcServer, stockService)
	return s.GrpcServer
}
