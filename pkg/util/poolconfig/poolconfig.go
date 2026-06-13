package poolconfig

import (
	"fmt"
	"runtime"
	"time"
)

// MySQLConfig MySQL连接池配置
type MySQLConfig struct {
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大存活时间
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
}

// RedisConfig Redis连接池配置
type RedisConfig struct {
	PoolSize        int           // 连接池大小
	MinIdleConns    int           // 最小空闲连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大存活时间
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
}

// GetMySQLConfig 获取推荐的MySQL连接池配置
// multiplier: 乘数，默认2，表示CPU核心数 * multiplier
func GetMySQLConfig(multiplier ...int) *MySQLConfig {
	cpuCount := runtime.NumCPU()
	m := 2
	if len(multiplier) > 0 && multiplier[0] > 0 {
		m = multiplier[0]
	}

	maxOpen := cpuCount * m
	maxIdle := maxOpen / 2
	if maxIdle < 5 {
		maxIdle = 5
	}

	return &MySQLConfig{
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// GetRedisConfig 获取推荐的Redis连接池配置
// multiplier: 乘数，默认5，表示CPU核心数 * multiplier
func GetRedisConfig(multiplier ...int) *RedisConfig {
	cpuCount := runtime.NumCPU()
	m := 5
	if len(multiplier) > 0 && multiplier[0] > 0 {
		m = multiplier[0]
	}

	poolSize := cpuCount * m
	minIdle := poolSize / 4
	if minIdle < 5 {
		minIdle = 5
	}
	maxIdle := poolSize / 2
	if maxIdle < 10 {
		maxIdle = 10
	}

	return &RedisConfig{
		PoolSize:        poolSize,
		MinIdleConns:    minIdle,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// GetCPUCount 获取CPU核心数
func GetCPUCount() int {
	return runtime.NumCPU()
}

// String 打印配置信息
func (c *MySQLConfig) String() string {
	return fmt.Sprintf("MySQLConfig{MaxOpenConns: %d, MaxIdleConns: %d, ConnMaxLifetime: %v, ConnMaxIdleTime: %v}",
		c.MaxOpenConns, c.MaxIdleConns, c.ConnMaxLifetime, c.ConnMaxIdleTime)
}

// String 打印Redis配置信息
func (c *RedisConfig) String() string {
	return fmt.Sprintf("RedisConfig{PoolSize: %d, MinIdleConns: %d, MaxIdleConns: %d, ConnMaxLifetime: %v, ConnMaxIdleTime: %v}",
		c.PoolSize, c.MinIdleConns, c.MaxIdleConns, c.ConnMaxLifetime, c.ConnMaxIdleTime)
}
