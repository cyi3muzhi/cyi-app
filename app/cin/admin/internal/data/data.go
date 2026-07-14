package data

import (
	"context"
	"fmt"

	productRpc "cyi-app/api/service/product/v1"
	userRpc "cyi-app/api/service/user/v1"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
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
)

// Data 包含所有下游微服务的 gRPC 客户端连接
type Data struct {
	pc  productRpc.ProductClient
	auc userRpc.AdminUserClient
}

// NewData .
func NewData(pc productRpc.ProductClient, auc userRpc.AdminUserClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the admin data resources")
	}
	return &Data{pc: pc, auc: auc}, cleanup, nil
}

// NewTokenValidator 创建 TokenValidator（wire provider）
// 通过 gRPC 调 user service 校验 token 是否有效（未被新登录覆盖）
func NewTokenValidator(auc userRpc.AdminUserClient) mid.TokenValidator {
	return func(ctx context.Context, rawToken string, claims *mid.UserClaims) error {
		reply, err := auc.ValidateToken(ctx, &userRpc.AdminValidateTokenRequest{Token: rawToken})
		if err != nil {
			return err
		}
		if !reply.Valid {
			return fmt.Errorf("token已失效，账号已在其他地方登录")
		}
		return nil
	}
}
