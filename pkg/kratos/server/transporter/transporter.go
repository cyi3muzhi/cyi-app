package transporter

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/bytedance/sonic"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/durationpb"
)

type TransporterOpt struct {
	Network string
	Addr    string
	Timeout *durationpb.Duration

	// 客户端服务发现
	ServiceName string
	Discovery   registry.Discovery
}

type MiddlewareOpt struct {
	// 认证 可选
	ApiKey         string
	ServiceKey     string
	WhiteList      []string
	TokenValidator mid.TokenValidator

	// Casbin 权限校验 可选
	Enforcer mid.EnforcerProvider

	// ExtraMiddlewares 附加中间件（合并进服务中间件链末尾执行，如操作日志）
	// 注意：不能通过第二个 http.Middleware() 追加，matcher.Use 是覆盖式赋值
	ExtraMiddlewares []middleware.Middleware

	// 链路追踪 *
	TracerProvider trace.TracerProvider
	// 日志 *
	Logger log.Logger
	// 指标 *
	Meter *mid.Meter
}

// http服务器选项
func NewHttpServerOpt(httpOpt *TransporterOpt, midOpt *MiddlewareOpt) []khttp.ServerOption {
	mids := []mid.MiddlewareOption{}
	if midOpt.TracerProvider == nil {
		panic("NewHttpServerOpt TracerProvider is nil")
	}
	if midOpt.Logger == nil {
		panic("NewHttpServerOpt Logger is nil")
	}
	if midOpt.Meter.Requests == nil || midOpt.Meter.Seconds == nil {
		panic("NewHttpServerOpt Meter is nil")
	}
	mids = append(mids, mid.WithLoggerMiddleware(midOpt.Logger))
	if midOpt.ApiKey != "" {
		mids = append(mids, mid.WithAuthMiddleware(mid.Auth{
			ApiKey:         midOpt.ApiKey,
			WhiteList:      midOpt.WhiteList,
			TokenValidator: midOpt.TokenValidator,
		}))
	}
	if midOpt.Enforcer != nil {
		mids = append(mids, mid.WithCasbinMiddleware(midOpt.Enforcer))
	}
	mids = append(mids, mid.WithMetricsMiddleware(midOpt.Meter))
	middlewares := mid.ServerMiddlewares(midOpt.TracerProvider, mids...)
	// 合并附加中间件（与主链同一个 khttp.Middleware，避免覆盖 matcher.defaults）
	if len(midOpt.ExtraMiddlewares) > 0 {
		middlewares = append(middlewares, midOpt.ExtraMiddlewares...)
	}

	opts := []khttp.ServerOption{
		khttp.Middleware(middlewares...),
		khttp.ResponseEncoder(khttp.DefaultResponseEncoder),
		khttp.ErrorEncoder(friendlyErrorEncoder),
		khttp.RequestDecoder(defaultRequestDecoder),
		khttp.Filter(handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins(corsAllowedOrigins()),
		)),
	}
	if httpOpt.Network != "" {
		opts = append(opts, khttp.Network(httpOpt.Network))
	}
	if httpOpt.Addr != "" {
		opts = append(opts, khttp.Address(httpOpt.Addr))
	}
	if httpOpt.Timeout != nil {
		opts = append(opts, khttp.Timeout(httpOpt.Timeout.AsDuration()))
	}
	return opts
}

// grpc服务器选项
func NewGrpcServerOpt(grpcOpt *TransporterOpt, midOpt *MiddlewareOpt) []grpc.ServerOption {
	mids := []mid.MiddlewareOption{}
	if midOpt.TracerProvider == nil {
		panic("NewGrpcServerOpt TracerProvider is nil")
	}
	if midOpt.Logger == nil {
		panic("NewGrpcServerOpt Logger is nil")
	}
	if midOpt.Meter.Requests == nil || midOpt.Meter.Seconds == nil {
		panic("NewGrpcServerOpt Meter is nil")
	}
	mids = append(mids, mid.WithLoggerMiddleware(midOpt.Logger))
	if midOpt.ApiKey != "" {
		mids = append(mids, mid.WithAuthMiddleware(mid.Auth{
			ApiKey: midOpt.ApiKey,
		}))
	}
	if midOpt.Enforcer != nil {
		mids = append(mids, mid.WithCasbinMiddleware(midOpt.Enforcer))
	}
	mids = append(mids, mid.WithMetricsMiddleware(midOpt.Meter))
	middlewares := mid.ServerMiddlewares(midOpt.TracerProvider, mids...)
	opts := []grpc.ServerOption{
		grpc.Middleware(middlewares...),
	}
	if grpcOpt.Network != "" {
		opts = append(opts, grpc.Network(grpcOpt.Network))
	}
	if grpcOpt.Addr != "" {
		opts = append(opts, grpc.Address(grpcOpt.Addr))
	}
	if grpcOpt.Timeout != nil {
		opts = append(opts, grpc.Timeout(grpcOpt.Timeout.AsDuration()))
	}
	return opts
}

// grpc客户端选项
func NewGrpcClientOpt(grpcOpt *TransporterOpt, midOpt *MiddlewareOpt) []grpc.ClientOption {
	if grpcOpt.ServiceName == "" {
		panic("grpcOpt.ServiceName is empty")
	}
	if grpcOpt.Discovery == nil {
		panic("grpcOpt.Discovery is nil")
	}
	if grpcOpt.Timeout == nil {
		panic("grpcOpt.Timeout is nil")
	}
	mids := []mid.MiddlewareOption{}
	if midOpt.TracerProvider == nil {
		panic("NewGrpcServerOpt TracerProvider is nil")
	}

	if midOpt.ServiceKey != "" {
		mids = append(mids, mid.WithAuthMiddleware(mid.Auth{
			ServiceKey: midOpt.ServiceKey,
		}))
	}
	if midOpt.Meter.Requests == nil || midOpt.Meter.Seconds == nil {
		panic("NewGrpcClientOpt Meter is nil")
	}
	mids = append(mids, mid.WithMetricsMiddleware(midOpt.Meter))
	middlewares := mid.ClientMiddlewares(midOpt.TracerProvider, mids...)
	opts := []grpc.ClientOption{
		grpc.WithEndpoint(fmt.Sprintf("discovery:///%v.%v", os.Getenv("SERVICE_ENV"), grpcOpt.ServiceName)),
		grpc.WithDiscovery(grpcOpt.Discovery),
		grpc.WithMiddleware(middlewares...),
		grpc.WithTimeout(grpcOpt.Timeout.AsDuration()),
	}
	return opts
}

type HttpServer struct {
	HttpServer *khttp.Server
}

type GrpcServer struct {
	GrpcServer *grpc.Server
}

type TaskServer struct {
	TaskServer *asynq.ServeMux
}

type Scheduler struct {
	Scheduler *asynq.Scheduler
}

func (s *TaskServer) CronHandleFunc(name string, f asynq.HandlerFunc) {
	s.TaskServer.HandleFunc(CronName(name), f)
}

func (s *TaskServer) AsyncHandleFunc(name string, f asynq.HandlerFunc) {
	s.TaskServer.HandleFunc(AsyncName(name), f)
}

func (s *TaskServer) DelayHandleFunc(name string, f asynq.HandlerFunc) {
	s.TaskServer.HandleFunc(DelayName(name), f)
}

// defaultRequestDecoder Kratos HTTP 请求解码器，Content-Type 为空时默认使用 JSON
func defaultRequestDecoder(r *http.Request, v interface{}) error {
	if r.Header.Get("Content-Type") == "" {
		r.Header.Set("Content-Type", "application/json")
	}
	return khttp.DefaultRequestDecoder(r, v)
}

// friendlyErrorEncoder Kratos HTTP 统一错误编码器：
//   - 业务错误（kratos error，含业务服务经 gRPC 编码传来的）：原样输出结构化 JSON
//     {code, reason, message}，前端可展示 message 中的友好文案
//   - 未知错误（网络、内部异常等，reason=Unknown）：替换为通用文案，避免把
//     "rpc error: ..." 等内部细节透传给前端
func friendlyErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := kerrors.FromError(err)
	if se == nil {
		se = kerrors.New(http.StatusInternalServerError, "INTERNAL_ERROR", "系统繁忙，请稍后重试")
	}
	// 非业务错误：不暴露内部错误细节
	if se.Reason == kerrors.UnknownReason {
		se.Reason = "INTERNAL_ERROR"
		se.Message = "系统繁忙，请稍后重试"
	}
	codec, _ := khttp.CodecForRequest(r, "Accept")
	body, merr := codec.Marshal(se)
	if merr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType(codec.Name()))
	if se.Code > 0 {
		w.WriteHeader(int(se.Code))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	_, _ = w.Write(body)
}

// contentType 将 codec 名称规范化为 HTTP Content-Type（application/json 等）
func contentType(name string) string {
	if strings.HasPrefix(name, "application/") {
		return name
	}
	return "application/" + name
}

// corsAllowedOrigins 返回 CORS 允许的跨域来源
// 优先从环境变量 CORS_ALLOWED_ORIGINS 读取（逗号分隔），便于部署时收紧；
// 未配置时默认允许所有来源，保证开发环境联调不受影响。
func corsAllowedOrigins() []string {
	if raw := os.Getenv("CORS_ALLOWED_ORIGINS"); raw != "" {
		parts := strings.Split(raw, ",")
		origins := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				origins = append(origins, p)
			}
		}
		if len(origins) > 0 {
			return origins
		}
	}
	return []string{"*"}
}

func CronName(name string) string {
	return fmt.Sprintf("%s:%s", "cron", name)
}

func AsyncName(name string) string {
	return fmt.Sprintf("%s:%s", "async", name)
}

func DelayName(name string) string {
	return fmt.Sprintf("%s:%s", "delay", name)
}

func NewAsyncTask(name string, payload interface{}, log *log.Helper) *asynq.Task {
	if payload == nil {
		payload = []byte{}
	}
	// 编码
	payloadBytes, err := sonic.Marshal(payload)
	if err != nil {
		log.Errorw("NewTask json.Marshal payload failed", "err", err)
		return nil
	}

	return asynq.NewTask(AsyncName(name), payloadBytes)
}

func NewDelayTask(name string, payload interface{}, log *log.Helper, delay time.Duration) *asynq.Task {
	if payload == nil {
		payload = []byte{}
	}
	// 编码
	payloadBytes, err := sonic.Marshal(payload)
	if err != nil {
		log.Errorw("NewTask json.Marshal payload failed", "err", err)
		return nil
	}
	if delay <= 0 {
		delay = time.Second
	}
	opts := []asynq.Option{
		asynq.ProcessIn(delay),
	}

	return asynq.NewTask(DelayName(name), payloadBytes, opts...)
}
