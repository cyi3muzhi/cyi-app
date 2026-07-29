package data

import (
	"cyi-app/app/rabbit/internal/conf"
	kdata "cyi-app/pkg/kratos/data"

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

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(c *conf.Data, logger log.Logger) *redis.Client {
	return kdata.NewRedisClient(
		c.Redis.Addr,
		c.Redis.Network,
		c.Redis.Password,
		c.Redis.ReadTimeout.AsDuration(),
		c.Redis.WriteTimeout.AsDuration(),
	)
}
