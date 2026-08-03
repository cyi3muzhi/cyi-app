package service

import (
	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
	"cyi-app/app/cin/admin/internal/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewAdminService,
)

// AdminService 实现了管理后台所有接口服务
// 嵌入多个 Unimplemented*Server 以兼容拆分后的 proto service
type AdminService struct {
	adminv1.UnimplementedDashboardServer
	adminv1.UnimplementedAdminProductServer
	adminv1.UnimplementedAdminOrderServer
	adminv1.UnimplementedAdminUserServer
	adminv1.UnimplementedSettingsServer
	adminv1.UnimplementedAdminAuthServer
	adminv1.UnimplementedAdminRoleServer
	adminv1.UnimplementedStaffServer
	adminv1.UnimplementedAdminCategoryServer
	adminv1.UnimplementedUploadServer
	adminv1.UnimplementedAdminCouponServer
	adminv1.UnimplementedBannerServer
	adminv1.UnimplementedAgentServer
	log       *log.Helper
	dc        *biz.DashboardUseCase
	pc        *biz.ProductUseCase
	ac        *biz.AuthUseCase
	rc        *biz.RoleUseCase
	sc        *biz.StaffUseCase
	cc        *biz.CategoryUseCase
	oc        *biz.OrderUseCase
	ec        *biz.ExportUseCase
	couponUC  *biz.CouponUseCase
	ur        *data.UploadRepo
	agentUC   *biz.AgentUseCase
	settUC    *biz.SettingsUseCase
	br        *data.BannerRepo
	userUC    *biz.UserUseCase
}

// NewAdminService 创建 AdminService
func NewAdminService(dc *biz.DashboardUseCase, pc *biz.ProductUseCase, ac *biz.AuthUseCase, rc *biz.RoleUseCase, sc *biz.StaffUseCase, cc *biz.CategoryUseCase, oc *biz.OrderUseCase, ec *biz.ExportUseCase, couponUC *biz.CouponUseCase, ur *data.UploadRepo, agentUC *biz.AgentUseCase, settingsUC *biz.SettingsUseCase, br *data.BannerRepo, userUC *biz.UserUseCase, logger log.Logger) *AdminService {
	return &AdminService{
		log:       log.NewHelper(logger),
		dc:        dc,
		pc:        pc,
		ac:        ac,
		rc:        rc,
		sc:        sc,
		cc:        cc,
		oc:        oc,
		ec:        ec,
		couponUC:  couponUC,
		ur:        ur,
		agentUC:   agentUC,
		settUC:    settingsUC,
		br:        br,
		userUC:    userUC,
	}
}
