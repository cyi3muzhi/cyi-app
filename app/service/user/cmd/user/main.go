package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cyi-app/app/service/user/internal/conf"
	"cyi-app/app/service/user/internal/data"

	cgrpc "cyi-app/pkg/kratos/client"
	consulConfig "cyi-app/pkg/kratos/config"

	tr "cyi-app/pkg/kratos/tracing"

	logutil "cyi-app/pkg/kratos/log"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/exporters/prometheus"
	_ "go.uber.org/automaxprocs"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	mid "cyi-app/pkg/kratos/server/middleware"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = fmt.Sprintf("%s.%s", os.Getenv("SERVICE_ENV"), cgrpc.USER_SERVICE)
	// Version is the version of the compiled software.
	Version string = cgrpc.USER_VERSION
	// 小写切去掉中间“-”
	id = strings.ReplaceAll(strings.ToLower(uuid.NewString()), "-", "")

	meteric *mid.Meter
)

func init() {

	exporter, err := prometheus.New()
	if err != nil {
		panic(err)
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	meter := provider.Meter(Name)

	_metricRequests, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		panic(err)
	}

	_metricSeconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		panic(err)
	}
	meteric = &mid.Meter{
		Requests: _metricRequests,
		Seconds:  _metricSeconds,
	}
}

func newApp(logger log.Logger, gs *grpc.Server, reg registry.Registrar, consumer *data.OperationLogConsumer) *kratos.App {
	// 启动 NATS 操作日志消费者（订阅 admin.operation 写库）
	go consumer.Start()
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
		),
		kratos.Registrar(reg),
		// 优雅停机钩子：先取消 NATS 订阅，等待在途消息处理完成，
		// 底层连接由 data.NewData 的 cleanup（nc.Close()）在资源清理阶段统一关闭。
		kratos.BeforeStop(func(ctx context.Context) error {
			if consumer != nil {
				consumer.Stop()
			}
			return nil
		}),
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
	// 创建TracerProvider
	tp := tr.NewTracerProvider(Name, bc.Trace.Endpoint)

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Auth, &cb, logger, tp, meteric)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
