package server

import (
	"cyi-app/app/task/internal/conf"
	consul "cyi-app/pkg/kratos/config"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/google/wire"

	ts "cyi-app/pkg/kratos/server/transporter"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewRegistrar,
	NewHTTPServerConf,
	NewHTTPServer,
	ProcessTask,
	NewServerMux,
	NewTaskServer,
	NewDiscovery,
)

func NewHTTPServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	metric *mid.Meter,
	logger log.Logger) *ts.HttpServer {
	srv := http.NewServer(ts.NewHttpServerOpt(
		&ts.TransporterOpt{
			Network: c.Http.Network,
			Addr:    c.Http.Addr,
			Timeout: c.Http.Timeout,
		},
		&ts.MiddlewareOpt{
			Logger:         logger,
			TracerProvider: tp,
			Meter:          metric,
		})...)
	return &ts.HttpServer{HttpServer: srv}
}

// 注册服务器
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	reg, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address: conf.Consul.Address,
		Scheme:  conf.Consul.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return reg
}

// 发现服务器
func NewDiscovery(conf *conf.Registry) registry.Discovery {
	dis, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address: conf.Consul.Address,
		Scheme:  conf.Consul.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return dis
}

// NewTaskServer 创建 asynq 服务端
// Queues 配置消费的队列及优先级：admin 投递订单导出任务到 "export" 队列，
// 必须在此声明消费，否则任务会堆积不被处理。
//
// 注意：不消费 "critical" 队列。该队列由 job 服务独占消费（cron 定时任务），
// 若 task 也消费 critical 队列会抢到 cron:user_count 等任务但因未注册 handler 而报错。
func NewTaskServer(r redis.UniversalClient, logger log.Logger) *asynq.Server {
	server := asynq.NewServerFromRedisClient(r, asynq.Config{
		Concurrency: 10,
		Logger:      log.NewHelper(logger),
		Queues: map[string]int{
			"default": 3, // 默认队列（email:send、create_user 等异步任务）
			"export":  2, // 导出队列（订单导出任务）
		},
	})
	return server
}

func NewServerMux(logger log.Logger) *ts.TaskServer {
	mux := asynq.NewServeMux()
	// 后期加个中间件
	// mux.Use()
	return &ts.TaskServer{TaskServer: mux}
}
