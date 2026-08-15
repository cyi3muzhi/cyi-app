package server

import (
	"cyi-app/app/service/order/internal/biz"
	pkgdtm "cyi-app/pkg/dtm"

	ts "cyi-app/pkg/kratos/server/transporter"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer 注册 DTM TCC 分支回调路由（在 NewHTTPServerConf 创建的 HTTP 服务器上）
// 路由路径与 biz 层编排的 URL 一致：/api/dtm/{branch}/{op}
func NewHTTPServer(s *ts.HttpServer, uc *biz.OrderUseCase) *khttp.Server {
	pkgdtm.RegisterTCCBranches(s.HttpServer, pkgdtm.TCCBasePath,
		pkgdtm.TCCBranch{
			Name:    "stock",
			Try:     uc.DtmStockTry,
			Confirm: uc.DtmStockConfirm,
			Cancel:  uc.DtmStockCancel,
		},
		pkgdtm.TCCBranch{
			Name:    "order",
			Try:     uc.DtmOrderTry,
			Confirm: uc.DtmOrderConfirm,
			Cancel:  uc.DtmOrderCancel,
		},
	)
	return s.HttpServer
}
