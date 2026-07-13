package service

import (
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	"cyi-app/app/cin/interface/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewCinInterface,
)

// CinInterface 实现了所有前台接口服务
// 嵌入多个 Unimplemented*Server 以兼容拆分后的 proto service
type CinInterface struct {
	v1.UnimplementedAuthServer
	v1.UnimplementedUserServer
	v1.UnimplementedProductServer
	v1.UnimplementedOrderServer
	v1.UnimplementedCartServer
	v1.UnimplementedStockServer
	v1.UnimplementedBladesServer
	log  *log.Helper
	uc   *biz.UserUseCase
	oc   *biz.OrderUseCase
	pc   *biz.ProductUseCase
	sc   *biz.StockUseCase
	bc   *biz.BladesUseCase
	auth *conf.Auth
}

func NewCinInterface(uc *biz.UserUseCase, oc *biz.OrderUseCase, pc *biz.ProductUseCase, sc *biz.StockUseCase, bc *biz.BladesUseCase, logger log.Logger, auth *conf.Auth) *CinInterface {
	return &CinInterface{
		log:  log.NewHelper(logger),
		uc:   uc,
		oc:   oc,
		pc:   pc,
		sc:   sc,
		bc:   bc,
		auth: auth,
	}
}
