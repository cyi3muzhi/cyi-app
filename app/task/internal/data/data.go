package data

import (
	"context"
	"cyi-app/app/task/internal/conf"
	"time"

	// "cyi-app/app/service/task/internal/data/curd"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewTaskRepo,
	NewData,
	NewRedisClient,
)

type Data struct {
	cache redis.UniversalClient
}

func NewData(c *conf.Data, cache redis.UniversalClient, logger log.Logger, server *asynq.Server) (*Data, func(), error) {
	cleanup := func() {
		server.Shutdown()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		cache: cache,
	}, cleanup, nil
}

func NewRedisClient(c *conf.Data, logger log.Logger) redis.UniversalClient {
	client := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Network:      c.Redis.Network,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		Password:     c.Redis.Password,
	})
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFunc()
	err := client.Ping(timeout).Err()
	if err != nil {
		panic(err)
	}
	return client
}
