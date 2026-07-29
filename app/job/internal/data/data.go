package data

import (
	userRpc "cyi-app/api/service/user/v1"
	marketingV1 "cyi-app/api/service/marketing/v1"
	"cyi-app/app/job/internal/conf"
	kdata "cyi-app/pkg/kratos/data"

	// "cyi-app/app/service/task/internal/data/curd"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewUserRepo,
	NewCouponRepo,
	NewJobRepo,
	NewData,
	NewRedisClient,
	NewUserServiceClient,
	NewMarketingCouponServiceClient,
)

type Data struct {
	userC   userRpc.UserClient
	couponC marketingV1.CouponServiceClient

	cache redis.UniversalClient
}

func NewData(c *conf.Data, cache redis.UniversalClient, logger log.Logger, server *asynq.Server, scheduler *asynq.Scheduler, userC userRpc.UserClient, couponC marketingV1.CouponServiceClient) (*Data, func(), error) {
	cleanup := func() {
		server.Shutdown()
		scheduler.Shutdown()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		userC:   userC,
		couponC: couponC,
		cache:   cache,
	}, cleanup, nil
}

func NewRedisClient(c *conf.Data, logger log.Logger) redis.UniversalClient {
	return kdata.NewRedisClient(
		c.Redis.Addr,
		c.Redis.Network,
		c.Redis.Password,
		c.Redis.ReadTimeout.AsDuration(),
		c.Redis.WriteTimeout.AsDuration(),
	)
}
