package server

import (
	"cyi-app/app/job/internal/conf"
	consul "cyi-app/pkg/kratos/config"
	kserver "cyi-app/pkg/kratos/server"
	"time"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/google/wire"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
	mid "cyi-app/pkg/kratos/server/middleware"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewRegistrar,
	NewDiscovery,
	NewHTTPServerConf,
	NewHTTPServer,
	ProcessTask,
	NewServerMux,
	NewTaskServer,
	NewScheduler,
	RegisterScheduler,
)

func NewHTTPServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	meteric *mid.Meter,
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
			Meter:          meteric,
		})...)
	return &ts.HttpServer{HttpServer: srv}
}

// 注册服务器
func NewRegistrar(c *conf.Registry) registry.Registrar {
	return kserver.NewRegistrar(c.Consul.Address, c.Consul.Scheme)
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

// NewTaskServer 创建 asynq 服务端，用于消费 cron 定时任务。
// job 服务作为 cron 调度器，既投递 cron 任务到 critical 队列，也消费该队列。
// 必须在此声明消费 critical 队列，否则投递的任务无人处理会被 task 服务误消费报错。
func NewTaskServer(r redis.UniversalClient, logger log.Logger) *asynq.Server {
	server := asynq.NewServerFromRedisClient(r, asynq.Config{
		Concurrency: 10,
		Logger:      log.NewHelper(logger),
		Queues: map[string]int{
			"critical": 5, // cron 任务专用队列（与 task 服务隔离）
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

func NewScheduler(r redis.UniversalClient, logger log.Logger) *ts.Scheduler {
	scheduler := asynq.NewSchedulerFromRedisClient(r, &asynq.SchedulerOpts{
		Logger: log.NewHelper(logger),
		Location: func() *time.Location {
			loc, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				// 加载失败时回退到UTC
				log.NewHelper(logger).Errorf("加载时区失败: %v, 使用默认UTC", err)
				return time.UTC
			}
			return loc
		}(),
	})
	return &ts.Scheduler{Scheduler: scheduler}
}

func RegisterScheduler(tasks []*conf.Task, scheduler *ts.Scheduler, mux *ts.TaskServer) *asynq.Scheduler {
	for _, task := range tasks {
		opts := make([]asynq.Option, 0)
		if task.Options != nil {
			if task.Options.MaxRetry != 0 {
				opts = append(opts, asynq.MaxRetry(int(task.Options.MaxRetry)))
			}
			if task.Options.Queue != "" {
				opts = append(opts, asynq.Queue(task.Options.Queue))
			}
			if task.Options.Timeout != nil {
				opts = append(opts, asynq.Timeout(task.Options.Timeout.AsDuration()))
			}
		}
		_, err := scheduler.Scheduler.Register(task.Cron, asynq.NewTask(ts.CronName(task.Name), nil), opts...)
		if err != nil {
			panic(err)
		}
	}
	return scheduler.Scheduler
}
