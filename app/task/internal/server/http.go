package server

import (
	ts "cyi-app/pkg/kratos/server/transporter"

	aqhelper "cyi-app/pkg/util/asynq_helper"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hibiken/asynqmon"
	"github.com/redis/go-redis/v9"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(s *ts.HttpServer, r redis.UniversalClient) *http.Server {
	s.HttpServer.HandlePrefix("/asynq/", asynqmon.New(asynqmon.Options{RedisConnOpt: aqhelper.NewQueueRedis(r)}))
	return s.HttpServer
}
