package server

import (
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/service"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	ts "cyi-app/pkg/kratos/server/transporter"
)

// NewGRPCServer new a gRPC server. 注册拆分后的多个接口服务
func NewGRPCServer(s *ts.GrpcServer, cinInterface *service.CinInterface) *grpc.Server {
	v1.RegisterAuthServer(s.GrpcServer, cinInterface)
	v1.RegisterUserServer(s.GrpcServer, cinInterface)
	v1.RegisterProductServer(s.GrpcServer, cinInterface)
	v1.RegisterOrderServer(s.GrpcServer, cinInterface)
	v1.RegisterCartServer(s.GrpcServer, cinInterface)
	v1.RegisterCategoryServer(s.GrpcServer, cinInterface)
	v1.RegisterStockServer(s.GrpcServer, cinInterface)
	v1.RegisterBladesServer(s.GrpcServer, cinInterface)
	v1.RegisterUserCouponServer(s.GrpcServer, cinInterface)
	return s.GrpcServer
}
