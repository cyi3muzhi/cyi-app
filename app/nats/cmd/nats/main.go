package main

import (
	"fmt"
	"os"
	"strings"

	"cyi-app/app/nats/internal/conf"
	"cyi-app/app/nats/internal/consumers"

	cgrpc "cyi-app/pkg/kratos/client"
	consulConfig "cyi-app/pkg/kratos/config"
	logutil "cyi-app/pkg/kratos/log"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/uuid"
	_ "go.uber.org/automaxprocs"
)

var (
	Name    string = fmt.Sprintf("%s.%s", os.Getenv("SERVICE_ENV"), cgrpc.NATS_SERVICE)
	Version string = cgrpc.NATS_VERSION
	id             = strings.ReplaceAll(strings.ToLower(uuid.NewString()), "-", "")
)

func newApp(logger log.Logger, reg registry.Registrar, consumer *consumers.Consumer) *kratos.App {
	go consumer.Start()
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
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
	// 创建zerolog适配器
	zerologAdapter := logutil.NewZerologAdapterWithConfig(&logutil.LogConfig{
		Level:               bc.Log.Level,
		EnableConsoleOutput: bc.Log.EnableConsole,
		EnableFileOutput:    bc.Log.EnableFile,
		LogDir:              bc.Log.LogDir,
		TimeFormat:          bc.Log.TimeFormat,
		EnableColor:         bc.Log.EnableColor,
		ServiceName:         Name,
	})

	// 创建增强版日志器
	logger := log.With(zerologAdapter,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Auth, bc.Nats, &cb, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
