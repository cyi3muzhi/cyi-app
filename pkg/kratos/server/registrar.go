package server

import (
	consul "cyi-app/pkg/kratos/config"

	"github.com/go-kratos/kratos/v2/registry"
)

// NewRegistrar 创建 Consul 服务注册器
func NewRegistrar(address, scheme string) registry.Registrar {
	reg, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address: address,
		Scheme:  scheme,
	})
	if err != nil {
		panic(err)
	}
	return reg
}
