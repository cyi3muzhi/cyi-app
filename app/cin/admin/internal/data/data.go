package data

import (
	"context"
	"fmt"

	adminv1 "cyi-app/api/cin/admin/v1"
	productRpc "cyi-app/api/service/product/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/conf"
	"cyi-app/pkg/util/rustfs"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDashboardRepo,
	NewProductRepo,
	NewAuthRepo,
	NewProductServiceClient,
	NewAdminUserServiceClient,
	NewTokenValidator,
	NewOssClient,
	NewRedisClient,
	NewCasbinEnforcer,
	ProvideEnforcerProvider,
)

// Data 包含所有下游微服务的 gRPC 客户端连接
type Data struct {
	pc    productRpc.ProductClient
	auc   userRpc.AdminUserClient
	oss   *rustfs.RustFSClient
	redis *redis.Client
}

// NewData .
func NewData(pc productRpc.ProductClient, auc userRpc.AdminUserClient, oss *rustfs.RustFSClient, redis *redis.Client, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the admin data resources")
	}
	return &Data{pc: pc, auc: auc, oss: oss, redis: redis}, cleanup, nil
}

// NewOssClient 创建 OSS 客户端
func NewOssClient(cfg *conf.Data) (*rustfs.RustFSClient, error) {
	return rustfs.NewRustFSClient(&rustfs.Config{
		Endpoint:       cfg.Oss.Endpoint,
		PublicEndpoint: cfg.Oss.PublicEndpoint,
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
