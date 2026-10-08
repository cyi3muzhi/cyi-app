package data

import (
	"fmt"
	"os"

	chatRpc "cyi-app/api/service/chat/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/interface/internal/conf"
	grpc "cyi-app/pkg/kratos/client"

	ts "cyi-app/pkg/kratos/server/transporter"

	marketingV1 "cyi-app/api/service/marketing/v1"
	orderRpc "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"
	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

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

// NewChatServiceClient 创建 IM 客服服务 gRPC 客户端（用户端会话/消息操作）
func NewChatServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) chatRpc.ChatClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.CHAT_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.CHAT_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.CHAT_SERVICE, err))
	}
	return chatRpc.NewChatClient(conn)
}

// NewMessageServiceClient 创建用户服务消息中心 gRPC 客户端
func NewMessageServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) userRpc.MessageClient {
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
	return userRpc.NewMessageClient(conn)
}

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

// 创建asynq客户端
func NewAsynqClient(r redis.UniversalClient) *asynq.Client {
	return asynq.NewClientFromRedisClient(r)
}

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

func NewMarketingCouponServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) marketingV1.CouponServiceClient {
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
	return marketingV1.NewCouponServiceClient(conn)
}

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

// NewAdminSettingsServiceClient 创建系统设置服务 gRPC 客户端
func NewAdminSettingsServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) userRpc.AdminSettingsClient {
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
	return userRpc.NewAdminSettingsClient(conn)
}

// NewAdminBannerServiceClient 创建 Banner 管理服务 gRPC 客户端
func NewAdminBannerServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) userRpc.AdminBannerClient {
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
	return userRpc.NewAdminBannerClient(conn)
}
