package data

import (
	"context"
	"fmt"
	"os"

	bladesRpc "cyi-app/api/blades/v1"
	adminv1 "cyi-app/api/cin/admin/v1"
	marketingRpc "cyi-app/api/service/marketing/v1"
	orderRpc "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/conf"
	"cyi-app/pkg/util/rustfs"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDashboardRepo,
	NewProductRepo,
	NewCategoryRepo,
	NewAuthRepo,
	NewStaffRepo,
	NewUploadRepo,
	NewOrderRepo,
	NewCouponRepo,
	NewSettingsRepo,
	NewBannerRepo,
	NewUserRepo,
	NewAdminBannerClient,
	NewProductServiceClient,
	NewCategoryServiceClient,
	NewAdminUserServiceClient,
	NewAdminRoleClient,
	NewAdminSettingsClient,
	NewOrderServiceClient,
	NewUserServiceClient,
	NewCouponServiceClient,
	NewBladesClient,
	NewTokenValidator,
	NewOssClient,
	NewRedisClient,
	ProvideUniversalClient,
	NewRoleRepo,
	NewRemoteEnforcer,
	ProvideEnforcerProvider,
	// 导出任务相关（仅投递，消费由 task 服务负责）
	NewExportRepo,
	NewAsynqClient,
)

// Data 包含所有下游微服务的 gRPC 客户端连接
type Data struct {
	pc          productRpc.ProductClient
	cc          productRpc.CategoryClient
	auc         userRpc.AdminUserClient
	oc          orderRpc.OrderClient
	uc          userRpc.UserClient
	mc          marketingRpc.CouponServiceClient
	bc          bladesRpc.BladesClient
	oss         *rustfs.RustFSClient
	redis       *redis.Client
	asynqClient *asynq.Client
}

// NewData 创建数据层实例
// admin 仅持有 asynq.Client 用于投递任务，不持有 asynq.Server；
// cleanup 时关闭 asynq 客户端连接。
func NewData(pc productRpc.ProductClient, cc productRpc.CategoryClient, auc userRpc.AdminUserClient, oc orderRpc.OrderClient, uc userRpc.UserClient, mc marketingRpc.CouponServiceClient, bc bladesRpc.BladesClient, oss *rustfs.RustFSClient, redis *redis.Client, asynqClient *asynq.Client, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		asynqClient.Close()
		log.NewHelper(logger).Info("closing the admin data resources")
	}
	return &Data{pc: pc, cc: cc, auc: auc, oc: oc, uc: uc, mc: mc, bc: bc, oss: oss, redis: redis, asynqClient: asynqClient}, cleanup, nil
}

// NewOssClient 创建 OSS 客户端
// PublicEndpoint 支持通过环境变量 OSS_PUBLIC_ENDPOINT 覆盖（部署环境用）
// 例如：OSS_PUBLIC_ENDPOINT=https://admin.cyi.dev/oss 走 Nginx 代理
func NewOssClient(cfg *conf.Data) (*rustfs.RustFSClient, error) {
	publicEndpoint := cfg.Oss.PublicEndpoint
	if env := os.Getenv("OSS_PUBLIC_ENDPOINT"); env != "" {
		publicEndpoint = env
	}
	return rustfs.NewRustFSClient(&rustfs.Config{
		Endpoint:       cfg.Oss.Endpoint,
		PublicEndpoint: publicEndpoint,
		AccessKey:      cfg.Oss.AccessKey,
		SecretKey:      cfg.Oss.SecretKey,
	})
}

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(cfg *conf.Data) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})
}

// NewTokenValidator 创建 TokenValidator（wire provider）
// 通过本地 Redis 校验 token 是否有效（未被新登录覆盖）
func NewTokenValidator(rdb *redis.Client) mid.TokenValidator {
	return func(ctx context.Context, rawToken string, claims *mid.UserClaims) error {
		storedToken, err := rdb.Get(ctx, fmt.Sprintf("admin:token:%d", claims.ID)).Result()
		if err != nil {
			if err == redis.Nil {
				return adminv1.ErrorUnauthorized("token已失效，账号已在其他地方登录")
			}
			return adminv1.ErrorUnauthorized("检查token失败")
		}
		if storedToken != rawToken {
			return adminv1.ErrorUnauthorized("token已失效，账号已在其他地方登录")
		}
		return nil
	}
}

// ProvideUniversalClient 将 *redis.Client 转换为 redis.UniversalClient（供 wire 注入）
func ProvideUniversalClient(client *redis.Client) redis.UniversalClient {
	return client
}
