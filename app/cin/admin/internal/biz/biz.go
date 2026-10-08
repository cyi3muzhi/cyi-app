package biz

import (
	"cyi-app/app/cin/admin/internal/conf"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewDashboardUseCase,
	NewProductUseCase,
	NewCategoryUseCase,
	NewAuthUseCase,
	NewRoleUseCase,
	NewStaffUseCase,
	NewOrderUseCase,
	NewExportUseCase,
	NewCouponUseCase,
	NewSettingsUseCase,
	NewAgentUseCase,
	NewUserUseCase,
	NewMessageUseCase,
	NewOperationLogUseCase,
	NewChatUseCase,
	ProvideApiKey,
)

// ProvideApiKey 提供 JWT 签名密钥（用于 wire 注入）
func ProvideApiKey(cfg *conf.Auth) string {
	return cfg.ApiKey
}
