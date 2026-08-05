package data

import (
	"fmt"
	"os"

	marketingRpc "cyi-app/api/service/marketing/v1"
	orderRpc "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/blades/internal/conf"
	grpc "cyi-app/pkg/kratos/client"

	ts "cyi-app/pkg/kratos/server/transporter"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/registry"
	"go.opentelemetry.io/otel/trace"
)

// NewProductServiceClient 创建商品服务 gRPC 客户端
func NewProductServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) productRpc.ProductClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.PRODUCT_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.PRODUCT_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.PRODUCT_SERVICE, err))
	}
	return productRpc.NewProductClient(conn)
}

// NewStockServiceClient 创建库存服务 gRPC 客户端（与 product 服务同连接）
func NewStockServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) productRpc.StockClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.PRODUCT_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.PRODUCT_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.PRODUCT_SERVICE, err))
	}
	return productRpc.NewStockClient(conn)
}

// NewUserServiceClient 创建用户服务 gRPC 客户端
func NewUserServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) userRpc.UserClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.USER_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.USER_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.USER_SERVICE, err))
	}
	return userRpc.NewUserClient(conn)
}

// NewAdminRoleClient 创建角色权限管理 gRPC 客户端（与 user 服务同连接）
func NewAdminRoleClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) userRpc.AdminRoleClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.USER_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.USER_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.USER_SERVICE, err))
	}
	return userRpc.NewAdminRoleClient(conn)
}

// NewOrderServiceClient 创建订单服务 gRPC 客户端
func NewOrderServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) orderRpc.OrderClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.ORDER_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.ORDER_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.ORDER_SERVICE, err))
	}
	return orderRpc.NewOrderClient(conn)
}

// NewMarketingServiceClient 创建营销（优惠券）服务 gRPC 客户端
func NewMarketingServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) marketingRpc.CouponServiceClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.MARKETING_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.MARKETING_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.MARKETING_SERVICE, err))
	}
	return marketingRpc.NewCouponServiceClient(conn)
}

// NewCategoryServiceClient 创建商品分类 gRPC 客户端（与 product 服务同连接）
func NewCategoryServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) productRpc.CategoryClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.PRODUCT_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.PRODUCT_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.PRODUCT_SERVICE, err))
	}
	return productRpc.NewCategoryClient(conn)
}
