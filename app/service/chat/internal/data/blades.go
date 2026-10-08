package data

import (
	"fmt"
	"os"

	bladesRpc "cyi-app/api/blades/v1"
	"cyi-app/app/service/chat/internal/conf"
	grpc "cyi-app/pkg/kratos/client"

	ts "cyi-app/pkg/kratos/server/transporter"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/registry"
	"go.opentelemetry.io/otel/trace"
)

// NewBladesClient 创建 Blades AI 服务 gRPC 客户端（M4 AI 客服兜底用）
func NewBladesClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) bladesRpc.BladesClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.BLADES_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.BLADES_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.BLADES_SERVICE, err))
	}
	return bladesRpc.NewBladesClient(conn)
}
