package data

import (
	"fmt"
	"os"

	orderRpc "cyi-app/api/service/order/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/task/internal/conf"
	grpc "cyi-app/pkg/kratos/client"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/registry"
	"go.opentelemetry.io/otel/trace"
	mid "cyi-app/pkg/kratos/server/middleware"
)

// NewUserServiceClient 创建用户服务 gRPC 客户端（查询用户名用于订单导出）
func NewUserServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, meter *mid.Meter) userRpc.UserClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.USER_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.USER_SERVICE],
			Meter:          meter,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.USER_SERVICE, err))
	}
	return userRpc.NewUserClient(conn)
}

// NewOrderServiceClient 创建订单服务 gRPC 客户端（订单导出时分页查询订单）
func NewOrderServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, meter *mid.Meter) orderRpc.OrderClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.ORDER_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.ORDER_SERVICE],
			Meter:          meter,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.ORDER_SERVICE, err))
	}
	return orderRpc.NewOrderClient(conn)
}
