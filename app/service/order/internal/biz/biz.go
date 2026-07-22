package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewOrderUseCase,
	NewOrderStatusHandler,
	NewWaitPayStatus,
	NewShipStatus,
	NewFinalStatus,
	NewCompleteStatus,
	NewPayStatus,
	NewRefundingStatus,
)
