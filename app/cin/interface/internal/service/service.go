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

type CinInterface struct {
	v1.UnimplementedCinInterfaceServer
	log  *log.Helper
	uc   *biz.UserUseCase
	oc   *biz.OrderUseCase
	pc   *biz.ProductUseCase
	auth *conf.Auth
}

func NewCinInterface(uc *biz.UserUseCase, oc *biz.OrderUseCase, pc *biz.ProductUseCase, logger log.Logger, auth *conf.Auth) *CinInterface {
	return &CinInterface{
		log:  log.NewHelper(logger),
		uc:   uc,
		oc:   oc,
		pc:   pc,
		auth: auth,
	}
}
