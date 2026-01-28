package data

import (
	"fmt"
	"os"

	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/interface/internal/conf"
	cgrpc "cyi-app/pkg/util/grpc"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/trace"
)

func NewUserServiceClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider) userRpc.UserClient {
	opts := []grpc.ClientOption{
		grpc.WithMiddleware(
			tracing.Client(tracing.WithTracerProvider(tp)),
			recovery.Recovery(),
			jwt.Client(func(token *jwt2.Token) (interface{}, error) {
				return []byte(auth.ServiceKey), nil
			},
				jwt.WithSigningMethod(jwt2.SigningMethodHS256),
			),
		),
		grpc.WithTimeout(conf.Grpc.Timeout.AsDuration()),
	}
	conn, err := cgrpc.ConnClient(cgrpc.USER_SERVICE, r, opts...)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), cgrpc.USER_SERVICE, err))
	}
	return userRpc.NewUserClient(conn)
}
