package data

import (
	"fmt"
	"os"

	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/service/chat/internal/conf"
	grpc "cyi-app/pkg/kratos/client"

	ts "cyi-app/pkg/kratos/server/transporter"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/registry"
	"go.opentelemetry.io/otel/trace"
)

// NewUserMessageClient 创建用户服务消息中心 gRPC 客户端（离线提醒发站内信用）
func NewUserMessageClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) userRpc.MessageClient {
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
