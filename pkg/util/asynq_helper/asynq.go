package asynqhelper

import (
	"fmt"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/redis/go-redis/v9"
)

// Client 是asynq客户端的封装
type Client struct {
	client *asynq.Client
}

// Server 是asynq服务器的封装
type Server struct {
	Server *asynq.Server
}

// NewClient 创建一个新的asynq客户端
// redisAddr: Redis服务器地址，格式为"host:port"
// redisPassword: Redis密码
// redisDB: Redis数据库编号
func NewClient(redisAddr, redisPassword string, redisDB int) (*Client, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// 测试连接
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Client{client: client}, nil
}

// NewServer 创建一个新的asynq服务器
// redisAddr: Redis服务器地址，格式为"host:port"
// redisPassword: Redis密码
// redisDB: Redis数据库编号
// concurrency: 并发处理任务的数量
func NewServer(r redis.UniversalClient, concurrency int, logger log.Logger) *Server {
	server := asynq.NewServer(
		newQueueRedis(r),
		asynq.Config{
			Concurrency: concurrency,
			Logger:      log.NewHelper(logger),
		},
	)

	return &Server{Server: server}
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.client.Close()
}

// Start 启动服务器
func (s *Server) Start(handler asynq.Handler) error {
	return s.Server.Start(handler)
}

// Stop 停止服务器
func (s *Server) Stop() {
	s.Server.Stop()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown() {
	s.Server.Shutdown()
}

func NewMonitor(r redis.UniversalClient) http.Handler {
	return asynqmon.New(asynqmon.Options{RedisConnOpt: newQueueRedis(r)})
}

type queueRedis struct {
	redis.UniversalClient
}

func newQueueRedis(c redis.UniversalClient) *queueRedis {
	return &queueRedis{
		UniversalClient: c,
	}
}

func (q *queueRedis) MakeRedisClient() interface{} {
	return q
}
