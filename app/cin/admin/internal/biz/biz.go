package biz

import (
	"cyi-app/app/cin/admin/internal/conf"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewDashboardUseCase,
	NewProductUseCase,
	NewAuthUseCase,
	NewRoleUseCase,
	ProvideApiKey,
)

func ProvideApiKey(cfg *conf.Auth) string {
	return cfg.ApiKey
}
