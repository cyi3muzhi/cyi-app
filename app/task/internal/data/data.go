package data

import (
	orderRpc "cyi-app/api/service/order/v1"
	userRpc "cyi-app/api/service/user/v1"
	marketingV1 "cyi-app/api/service/marketing/v1"
	"cyi-app/app/task/internal/biz"
	"cyi-app/app/task/internal/conf"
	kdata "cyi-app/pkg/kratos/data"
	"cyi-app/pkg/util/rustfs"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
// 包含：用户/订单/营销 gRPC 客户端、OSS 客户端、Redis、asynq、各业务 repo。
var ProviderSet = wire.NewSet(
	NewUserRepo,
	NewTaskRepo,
	NewCouponRepo,
	NewData,
	NewRedisClient,
	NewUserServiceClient,
	NewOrderServiceClient,
	NewMarketingCouponServiceClient,
	NewOssClient,
	// 订单导出相关
	NewExportRepo,
	NewExcelGenerator,
	NewOrderRepo,
	// 订单超时自动取消
	NewOrderTimeoutRepo,
	// 接口绑定：ExcelGenerator 实现 biz.ExcelWriter 接口
	wire.Bind(new(biz.ExcelWriter), new(*ExcelGenerator)),
)

// Data 包含下游微服务 gRPC 客户端与基础资源
type Data struct {
	userC   userRpc.UserClient
	oc      orderRpc.OrderClient    // 订单服务客户端（导出查询用）
	couponC marketingV1.CouponServiceClient // 营销服务客户端（优惠券批量过期 cron 任务用）
	oss     *rustfs.RustFSClient    // OSS 客户端（导出文件上传用）
	cache   redis.UniversalClient
}

// NewData 创建数据层实例
// server 用于 cleanup 时关闭 asynq 服务端
func NewData(c *conf.Data, cache redis.UniversalClient, logger log.Logger, server *asynq.Server, userC userRpc.UserClient, oc orderRpc.OrderClient, couponC marketingV1.CouponServiceClient, oss *rustfs.RustFSClient) (*Data, func(), error) {
	cleanup := func() {
		server.Shutdown()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		cache:   cache,
		userC:   userC,
		oc:      oc,
		couponC: couponC,
		oss:     oss,
	}, cleanup, nil
}

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(c *conf.Data, logger log.Logger) redis.UniversalClient {
	return kdata.NewRedisClient(
		c.Redis.Addr,
		c.Redis.Network,
		c.Redis.Password,
		c.Redis.ReadTimeout.AsDuration(),
		c.Redis.WriteTimeout.AsDuration(),
	)
}

// NewOssClient 创建 OSS（RustFS）客户端，用于订单导出文件上传
func NewOssClient(cfg *conf.Data) (*rustfs.RustFSClient, error) {
	return rustfs.NewRustFSClient(&rustfs.Config{
		Endpoint:       cfg.Oss.Endpoint,
		PublicEndpoint: cfg.Oss.PublicEndpoint,
		AccessKey:      cfg.Oss.AccessKey,
		SecretKey:      cfg.Oss.SecretKey,
	})
}
