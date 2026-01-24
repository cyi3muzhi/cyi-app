package service

import (
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewCinInterface,
)

type CinInterface struct {
	v1.UnimplementedCinInterfaceServer
	log *log.Helper
	uc  *biz.UserUseCase
	ac  *biz.AuthUseCase
}

func NewCinInterface(uc *biz.UserUseCase, ac *biz.AuthUseCase, logger log.Logger) *CinInterface {
	return &CinInterface{
		log: log.NewHelper(logger),
		uc:  uc,
		ac:  ac,
	}
}
