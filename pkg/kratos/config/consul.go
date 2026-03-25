package config

import (
	"fmt"
	"os"
	"strings"

	consulConfig "github.com/go-kratos/kratos/contrib/config/consul/v2"
	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/hashicorp/consul/api"
)

// ConsulConfigOptions 定义Consul配置选项
type ConsulConfigOptions struct {
	Address string
	Path    string
	Paths   []string // 多配置路径
}

// 获取consul配置
func NewConsulConfig(opts *ConsulConfigOptions) (config.Config, error) {
	// 设置默认值
	address := opts.Address
	if address == "" {
		address = os.Getenv("CONSUL_ADDR")
	}

	paths := opts.Paths
	if paths == nil {
		paths = strings.Split(os.Getenv("CONSUL_PATHS"), ",")
		for i,path := range paths{
			paths[i] = strings.TrimSpace(path)
		}
	}

	client, err := api.NewClient(&api.Config{
		Address: address,
	})
	if err != nil {
		return nil, err
	}
	sources := make([]config.Source, 0)
	for _, path := range paths {
		source, err := consulConfig.New(client, consulConfig.WithPath(path))
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}

	c := config.New(
		config.WithSource(sources...),
	)

	return c, nil
}

// ConsulRegistryOptions 定义Consul注册中心选项
type ConsulRegistryOptions struct {
	Address           string
	Scheme            string
	Token             string
	Username          string
	Password          string
	EnableHealthCheck bool
	CheckInterval     int32
}

// NewConsulRegistry 创建Consul注册中心客户端
func NewConsulRegistry(opts *ConsulRegistryOptions) (*consul.Registry, error) {
	// 参数验证
	if opts == nil {
		return nil, fmt.Errorf("consul registry options cannot be nil")
	}

	// 设置默认值
	address := opts.Address
	if address == "" {
		address = os.Getenv("CONSUL_ADDR")
		if address == "" {
			return nil, fmt.Errorf("consul address is empty, set it in options or CONSUL_ADDR environment variable")
		}
	}

	scheme := opts.Scheme
	if scheme == "" {
		scheme = "http"
	}

	// 创建配置并初始化HttpAuth结构体防止nil指针panic
	config := &api.Config{
		Address: address,
		Scheme:  scheme,
	}

	// 如果设置了用户名或密码，初始化HttpAuth
	if opts.Username != "" || opts.Password != "" {
		config.HttpAuth = &api.HttpBasicAuth{}
		config.HttpAuth.Username = opts.Username
		config.HttpAuth.Password = opts.Password
	}

	// 设置Token（如果有）
	if opts.Token != "" {
		config.Token = opts.Token
	}

	// 创建Consul客户端
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	// 配置Consul选项
	consulOpts := []consul.Option{
		consul.WithHealthCheck(opts.EnableHealthCheck),
	}

	// 修复健康检查间隔逻辑：只有当间隔大于0时才添加
	if opts.CheckInterval > 0 {
		consulOpts = append(consulOpts, consul.WithHealthCheckInterval(int(opts.CheckInterval)))
	}

	// 创建并返回Registry实例
	return consul.New(client, consulOpts...), nil
}
