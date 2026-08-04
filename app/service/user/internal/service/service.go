package service

import (
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewUserService,
	NewAdminUserService,
	NewAdminRoleService,
	NewAdminSettingsService,
	NewAdminBannerService,
	NewOperationLogService,
)
