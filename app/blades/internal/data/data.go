package data

import (
	"context"
	"fmt"
	"time"

	orderRpc "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/blades/internal/biz"
	"cyi-app/app/blades/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
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

// Data 持有 Redis 连接和各下游服务的 gRPC 客户端
type Data struct {
	cache *redis.Client
	pc    productRpc.ProductClient
	uc    userRpc.UserClient
	oc    orderRpc.OrderClient
}

// NewData .
func NewData(cache *redis.Client, pc productRpc.ProductClient, uc userRpc.UserClient, oc orderRpc.OrderClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("redis client is nil")
	}
	return &Data{
		cache: cache,
		pc:    pc,
		uc:    uc,
		oc:    oc,
	}, cleanup, nil
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
