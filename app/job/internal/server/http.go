package server

import (
	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hibiken/asynqmon"
	"github.com/redis/go-redis/v9"
)

// queueRedis 适配 redis.UniversalClient 到 asynq.RedisConnOpt 接口
type queueRedis struct {
	redis.UniversalClient
}

func (q *queueRedis) MakeRedisClient() interface{} {
	return q
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(s *ts.HttpServer, r redis.UniversalClient) *http.Server {
	s.HttpServer.HandlePrefix("", asynqmon.New(asynqmon.Options{RedisConnOpt: &queueRedis{UniversalClient: r}}))
	return s.HttpServer
}
