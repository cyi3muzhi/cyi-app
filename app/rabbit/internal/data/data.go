package data

import (
	"context"
	"cyi-app/app/rabbit/internal/conf"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewRedisClient,
	NewOrderRepo,
)

type Data struct {
	cache *redis.Client
}

func NewData(c *conf.Data, cache *redis.Client, logger log.Logger) (*Data, func(), error) {

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	// curd.SetDefault(db)
	return &Data{
		cache: cache,
	}, cleanup, nil
}

func NewRedisClient(c *conf.Data, logger log.Logger) *redis.Client {
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
