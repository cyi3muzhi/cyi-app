package biz

import (
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
	ProvideApiKey,
)

func ProvideApiKey(apiKey string) string {
	return apiKey
}
