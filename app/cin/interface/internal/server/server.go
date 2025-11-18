package server

import (
	"cyi-app/app/cin/interface/internal/conf"
	consul "cyi-app/pkg/util/config"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar, NewDiscovery)

// 注册服务器
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	reg, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address:           conf.Consul.Address,
		Scheme:            conf.Consul.Scheme,
		EnableHealthCheck: conf.Consul.EnableHealthCheck,
		CheckInterval:     conf.Consul.CheckInterval,
	})
	if err != nil {
		panic(err)
	}
	return reg
}

// 发现服务器
func NewDiscovery(conf *conf.Registry) registry.Discovery {
	dis, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address: conf.Consul.Address,
		Scheme:  conf.Consul.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return dis
}
