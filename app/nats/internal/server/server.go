package server

import (
	"cyi-app/app/nats/internal/conf"
	consul "cyi-app/pkg/kratos/config"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewRegistrar, NewNatsConn)

// NewRegistrar 注册服务器
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	reg, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address: conf.Consul.Address,
		Scheme:  conf.Consul.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return reg
}

// NewNatsConn 创建NATS连接
func NewNatsConn(nc *conf.Nats) *nats.Conn {
	conn, err := nats.Connect(nc.Url)
	if err != nil {
		panic(err)
	}
	return conn
}
