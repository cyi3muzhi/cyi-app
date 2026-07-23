package data

import (
	"fmt"
	"os"

	"cyi-app/app/service/order/internal/conf"
	grpc "cyi-app/pkg/kratos/client"

	ts "cyi-app/pkg/kratos/server/transporter"

	productRpc "cyi-app/api/service/product/v1"
	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/registry"
	"go.opentelemetry.io/otel/trace"
)

func NewProductServiceClient(r registry.Discovery, auth *conf.Auth, conf *conf.Server, tp trace.TracerProvider, metric *mid.Meter) productRpc.ProductClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.PRODUCT_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			Meter:          metric,
			ServiceKey:     auth.ServiceKey[grpc.PRODUCT_SERVICE],
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.PRODUCT_SERVICE, err))
	}
	return productRpc.NewProductClient(conn)
}


