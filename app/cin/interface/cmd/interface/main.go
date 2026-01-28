package main

import (
	"context"
	"fmt"
	"os"

	"cyi-app/app/cin/interface/internal/conf"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	consulConfig "cyi-app/pkg/util/config"
	cgrpc "cyi-app/pkg/util/grpc"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = fmt.Sprintf("%s.%s", os.Getenv("SERVICE_ENV"), cgrpc.CIN_INTERFACE)
	// Version is the version of the compiled software.
	Version string = cgrpc.CIN_VERSION

	id = uuid.NewString()
)

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, reg registry.Registrar) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
		kratos.Registrar(reg),
	)
}

func main() {
	// consul配置初始化
	c, err := consulConfig.NewConsulConfig(&consulConfig.ConsulConfigOptions{})
	if err != nil {
		panic(err)
	}

	defer c.Close()

	// 加载配置
	if err = c.Load(); err != nil {
		panic(err)
	}

	// 扫描配置
	var bc conf.Bootstrap
	if err = c.Scan(&bc); err != nil {
		panic(err)
	}
	var cb conf.Registry
	if err = c.Scan(&cb); err != nil {
		panic(err)
	}

	var awl conf.AuthWhiteList
	if err = c.Scan(&awl); err != nil {
		panic(err)
	}
	fmt.Println("awl:", awl)
	var tl conf.TransactionList
	if err = c.Scan(&tl); err != nil {
		panic(err)
	}
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(bc.Trace.Endpoint),
		otlptracehttp.WithInsecure(), // 允许HTTP连接，不强制要求HTTPS
	)
	if err != nil {
		panic(err)
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(Name),
		)),
	)

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Auth, &cb, logger, tp, &awl, &tl)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
