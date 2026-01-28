package data

import (
	"fmt"
	"os"

	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/interface/internal/conf"
	cgrpc "cyi-app/pkg/util/grpc"

	mid "cyi-app/pkg/util/server/middleware"
	ts "cyi-app/pkg/util/server/transporter"

	"github.com/go-kratos/kratos/v2/registry"
	"go.opentelemetry.io/otel/trace"
)

func NewUserServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider) userRpc.UserClient {
	midList := mid.ClientMiddlewares(tp,
		mid.WithAuthMiddleware(mid.Auth{
			ServiceKey: auth.ServiceKey,
		}),
	)
	opts := ts.NewGrpcClientOpt(
		ts.TransporterOpt{
			Timeout: conf.Grpc.Timeout,
		},
		midList...,
	)
	conn, err := cgrpc.ConnClient(cgrpc.USER_SERVICE, r, opts...)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), cgrpc.USER_SERVICE, err))
	}
	return userRpc.NewUserClient(conn)
}
