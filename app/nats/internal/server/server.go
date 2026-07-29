package server

import (
	"cyi-app/app/nats/internal/conf"
	kserver "cyi-app/pkg/kratos/server"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewRegistrar, NewNatsConn)

// NewRegistrar 注册服务器
func NewRegistrar(c *conf.Registry) registry.Registrar {
	return kserver.NewRegistrar(c.Consul.Address, c.Consul.Scheme)
}

// NewNatsConn 创建NATS连接
func NewNatsConn(nc *conf.Nats) *nats.Conn {
	conn, err := nats.Connect(nc.Url)
	if err != nil {
		panic(err)
	}
	return conn
}
