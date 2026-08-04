package biz

import (
	"cyi-app/app/service/user/internal/conf"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewUserUseCase,
	NewAddressUseCase,
	NewAdminUserUseCase,
	NewRoleUseCase,
	NewSettingsUseCase,
	NewBannerUseCase,
	NewOperationLogUseCase,
	ProvideApiKey,
)

// ProvideApiKey 提供 API Key（用于 wire 注入）
func ProvideApiKey(cfg *conf.Auth) string {
	return cfg.ApiKey
}
