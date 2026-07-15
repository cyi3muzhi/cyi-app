package rustfs

import (
	"fmt"
	"os"
)

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	Endpoint:  "http://localhost:9000",
	AccessKey: "rustfsadmin",
	SecretKey: "rustfsadmin",
}

// InitDefaultClient 初始化默认RustFS客户端
func InitDefaultClient() (*RustFSClient, error) {
	return NewRustFSClient(DefaultConfig)
}

// InitFromEnv 从环境变量初始化RustFS客户端
func InitFromEnv() (*RustFSClient, error) {
	cfg := &Config{
		Endpoint:  os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey: os.Getenv("RUSTFS_ACCESS_KEY"),
		SecretKey: os.Getenv("RUSTFS_SECRET_KEY"),
	}

	// 验证环境变量
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("missing RUSTFS_ENDPOINT environment variable")
	}
	if cfg.AccessKey == "" {
		return nil, fmt.Errorf("missing RUSTFS_ACCESS_KEY environment variable")
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("missing RUSTFS_SECRET_KEY environment variable")
	}

	return NewRustFSClient(cfg)
}

// InitWithConfig 使用自定义配置初始化RustFS客户端
func InitWithConfig(endpoint, accessKey, secretKey string) (*RustFSClient, error) {
	cfg := &Config{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	return NewRustFSClient(cfg)
}
