package server

import (
	"cyi-app/app/service/marketing/internal/service"

	v1 "cyi-app/api/service/marketing/v1"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewServer(s *ts.GrpcServer, couponService *service.CouponService,
) *grpc.Server {
	v1.RegisterCouponServiceServer(s.GrpcServer, couponService)
	return s.GrpcServer
}