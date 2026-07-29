package data

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient 创建 Redis 客户端并进行连通性检查
func NewRedisClient(addr, network, password string, readTimeout, writeTimeout time.Duration) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Network:      network,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Password:     password,
	})
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFunc()
	err := rdb.Ping(timeout).Err()
	if err != nil {
		panic(err)
	}
	return rdb
}
