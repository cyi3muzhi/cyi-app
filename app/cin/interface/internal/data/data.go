package data

import (
	"context"
	"os"
	"time"

	"cyi-app/app/cin/interface/internal/conf"

	"cyi-app/pkg/util/rustfs"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	chatRpc "cyi-app/api/service/chat/v1"
	marketingV1 "cyi-app/api/service/marketing/v1"
	orderRpc "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"
	userRpc "cyi-app/api/service/user/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewUserRepo,
	NewUserServiceClient,
	NewMessageRepo,
	NewMessageServiceClient,
	NewOrderRepo,
	NewOrderServiceClient,
	NewProductRepo,
	NewProductServiceClient,
	NewCategoryRepo,
	NewCategoryServiceClient,
	NewStockRepo,
	NewStockServiceClient,
	NewCouponRepo,
	NewMarketingCouponServiceClient,
	NewSettingsRepo,
	NewAdminSettingsServiceClient,
	NewBannerRepo,
	NewAdminBannerServiceClient,
	NewUploadRepo,
	NewCartRepo,
	NewRedisClient,
	NewAsynqClient,
	NewOssClient,
	NewChatServiceClient,
	NewChatRepo,
	NewNATSConn,
)

// Data .
type Data struct {
	uc  userRpc.UserClient
	mc  userRpc.MessageClient
	oc  orderRpc.OrderClient
	pc  productRpc.ProductClient
	cc  productRpc.CategoryClient
	sc  productRpc.StockClient
	mcc marketingV1.CouponServiceClient
	asc userRpc.AdminSettingsClient
	abc userRpc.AdminBannerClient
	chc chatRpc.ChatClient
	rc  redis.UniversalClient
	mq  *asynq.Client
	oss *rustfs.RustFSClient
	nc  *nats.Conn
}

// NewData .
func NewData(c *conf.Data, uc userRpc.UserClient, mc userRpc.MessageClient, oc orderRpc.OrderClient, pc productRpc.ProductClient, cc productRpc.CategoryClient, sc productRpc.StockClient, mcc marketingV1.CouponServiceClient, asc userRpc.AdminSettingsClient, abc userRpc.AdminBannerClient, chc chatRpc.ChatClient, rc redis.UniversalClient, logger log.Logger, mq *asynq.Client, oss *rustfs.RustFSClient, nc *nats.Conn) (*Data, func(), error) {
	cleanup := func() {
		if nc != nil {
			nc.Close()
		}
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{uc: uc, mc: mc, oc: oc, pc: pc, cc: cc, sc: sc, mcc: mcc, asc: asc, abc: abc, chc: chc, rc: rc, mq: mq, oss: oss, nc: nc}, cleanup, nil
}

// NewOssClient 创建 OSS 客户端（用户服务 bucket）
// PublicEndpoint 支持通过环境变量 OSS_PUBLIC_ENDPOINT 覆盖（部署环境走 Nginx 代理）
func NewOssClient(cfg *conf.Data) (*rustfs.RustFSClient, error) {
	publicEndpoint := cfg.Oss.GetPublicEndpoint()
	if env := os.Getenv("OSS_PUBLIC_ENDPOINT"); env != "" {
		publicEndpoint = env
	}
	return rustfs.NewRustFSClient(&rustfs.Config{
		Endpoint:       cfg.Oss.GetEndpoint(),
		PublicEndpoint: publicEndpoint,
		AccessKey:      cfg.Oss.GetAccessKey(),
		SecretKey:      cfg.Oss.GetSecretKey(),
	})
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
