package data

import (
	"context"
	"time"

	"cyi-app/app/cin/interface/internal/conf"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	userRpc "cyi-app/api/service/user/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewUserRepo,
	NewUserServiceClient,
	NewRedisClient,
	NewAsynqClient,
)

// Data .
type Data struct {
	uc userRpc.UserClient
	rc redis.UniversalClient
	mq *asynq.Client
}

// NewData .
func NewData(c *conf.Data, uc userRpc.UserClient, rc redis.UniversalClient, logger log.Logger, mq *asynq.Client) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{uc: uc, rc: rc, mq: mq}, cleanup, nil
}

// NewRedisClient .
func NewRedisClient(c *conf.Data, logger log.Logger) redis.UniversalClient {
	client := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Network:      c.Redis.Network,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		Password:     c.Redis.Password,
	})
	// client.AddHook(redisotel.NewTracingHook())
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFunc()
	err := client.Ping(timeout).Err()
	if err != nil {
		panic(err)
	}
	return client
}
