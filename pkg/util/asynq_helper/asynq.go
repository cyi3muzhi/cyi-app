package asynqhelper

import (
	"fmt"

	"github.com/hibiken/asynq"
)

// Client 是asynq客户端的封装
type Client struct {
	client *asynq.Client
}

// Server 是asynq服务器的封装
type Server struct {
	server *asynq.Server
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
func NewServer(redisAddr, redisPassword string, redisDB int, concurrency int) (*Server, error) {
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDB,
		},
		asynq.Config{
			Concurrency: concurrency,
		},
	)

	return &Server{server: server}, nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.client.Close()
}

// Start 启动服务器
func (s *Server) Start(handler asynq.Handler) error {
	return s.server.Start(handler)
}

// Stop 停止服务器
func (s *Server) Stop() {
	s.server.Stop()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown() {
	s.server.Shutdown()
}
