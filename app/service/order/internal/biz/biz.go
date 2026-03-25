package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewOrderUseCase,
	NewOrderStatusHandler,
	NewWaitPayStatus,
	NewShipStatus,
	NewCancelStatus,
	NewRefundedStatus,
	NewCompleteStatus,
	NewPayStatus,
	NewRefundingStatus,
)
