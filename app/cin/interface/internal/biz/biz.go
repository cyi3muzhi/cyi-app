package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewUserUseCase,
	NewMessageUseCase,
	NewOrderUseCase,
	NewProductUseCase,
	NewCategoryUseCase,
	NewStockUseCase,
	NewCouponUseCase,
	NewBladesUseCase,
	NewSettingsUseCase,
	NewBannerUseCase,
	NewCartUseCase,
)
