package data

import (
	"context"
	"time"

	"cyi-app/app/blades/internal/biz"
	"cyi-app/app/blades/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewRedisClient,
	ProvideUniversalClient,
	NewVectorStore,
	ProviderVectorStoreDimension,
	wire.Bind(new(biz.VectorStore), new(*VectorStore)),
	NewProductServiceClient,
	NewStockServiceClient,
	NewUserServiceClient,
	NewAdminRoleClient,
	NewOrderServiceClient,
)

// ProviderVectorStoreDimension 从配置获取向量维度
func ProviderVectorStoreDimension(cfg *conf.Blades) int {
	if cfg != nil && cfg.Embedding != nil && cfg.Embedding.Dimension > 0 {
		return int(cfg.Embedding.Dimension)
	}
	return 768
}

// NewRedisClient .
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
		log.NewHelper(logger).Warnf("redis ping failed (addr=%s): %v, will retry on each operation", c.Redis.Addr, err)
	}
	return client
}

// ProvideUniversalClient 将 *redis.Client 转换为 redis.UniversalClient
func ProvideUniversalClient(client *redis.Client) redis.UniversalClient {
	return client
}
