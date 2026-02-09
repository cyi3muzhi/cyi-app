package service

import (
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	"cyi-app/app/cin/interface/internal/conf"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/gorilla/websocket"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewCinInterface,
)

type CinInterface struct {
	v1.UnimplementedCinInterfaceServer
	log  *log.Helper
	uc   *biz.UserUseCase
	ac   *biz.AuthUseCase
	auth *conf.Auth

	// 连接管理
	userConnections map[int64]*websocket.Conn
	groupMembers    map[int64][]int64
	mu              sync.RWMutex
}

func NewCinInterface(uc *biz.UserUseCase, ac *biz.AuthUseCase, logger log.Logger, auth *conf.Auth) *CinInterface {
	return &CinInterface{
		log:  log.NewHelper(logger),
		uc:   uc,
		ac:   ac,
		auth: auth,

		// 初始化连接管理
		userConnections: make(map[int64]*websocket.Conn),
		groupMembers:    make(map[int64][]int64),
	}
}
