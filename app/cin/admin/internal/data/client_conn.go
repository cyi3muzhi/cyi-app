package data

import (
	"fmt"
	"os"

	productRpc "cyi-app/api/service/product/v1"
	"cyi-app/app/cin/admin/internal/conf"
	grpc "cyi-app/pkg/kratos/client"

	ts "cyi-app/pkg/kratos/server/transporter"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/registry"
	"go.opentelemetry.io/otel/trace"
)

// NewProductServiceClient 创建商品服务 gRPC 客户端连接
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
