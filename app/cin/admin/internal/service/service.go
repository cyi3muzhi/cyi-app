package service

import (
	v1 "cyi-app/api/cin/admin/user"
	"cyi-app/app/cin/admin/internal/biz"
	"cyi-app/app/cin/admin/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewCinAdmin,
)

type CinAdmin struct {
	v1.UnimplementedUserServiceServer
	log  *log.Helper
	uc   *biz.UserUseCase
	ac   *biz.AuthUseCase
	auth *conf.Auth
}

func NewCinAdmin(uc *biz.UserUseCase, ac *biz.AuthUseCase, logger log.Logger, auth *conf.Auth) *CinAdmin {
	return &CinAdmin{
		log:  log.NewHelper(logger),
		uc:   uc,
		ac:   ac,
		auth: auth,
	}
}
