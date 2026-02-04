package cin_redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	OneDay = time.Hour * 24
	OneWeek = time.Hour * 24 * 7
	OneMonth = time.Hour * 24 * 30
	OneYear = time.Hour * 24 * 365
	KeepTTL = redis.KeepTTL
)

// string
func SetRedis(ctx context.Context, rc *redis.Client, key interface{}, value interface{}, expiration time.Duration) error {
	keyStr := fmt.Sprintf("%v", key)
	if valueStr, ok := value.(string); ok {
		return rc.Set(ctx, keyStr, valueStr, expiration).Err()
	}
	// 序列化值为 JSON 字符串
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	fmt.Println("SetRedis:", keyStr, string(jsonValue))
	return rc.Set(ctx, keyStr, jsonValue, expiration).Err()
}

func GetRedis(ctx context.Context, rc *redis.Client, key interface{}, value interface{}) error {
	keyStr := fmt.Sprintf("%v", key)
	jsonValue, err := rc.Get(ctx, keyStr).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonValue), value)
}

func DelRedis(ctx context.Context, rc *redis.Client, key interface{}) error {
	keyStr := fmt.Sprintf("%v", key)
	return rc.Del(ctx, keyStr).Err()
}

// hash
func SetHashRedis(ctx context.Context, rc *redis.Client, key interface{}, field interface{}, value interface{}, expiration time.Duration) error {
	keyStr := fmt.Sprintf("%v", key)
	fieldStr := fmt.Sprintf("%v", field)
	if valueStr, ok := value.(string); ok {
		return rc.HSet(ctx, keyStr, fieldStr, valueStr).Err()
	}
	// 序列化值为 JSON 字符串
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rc.HSet(ctx, keyStr, fieldStr, jsonValue).Err()
}

func GetHashRedis(ctx context.Context, rc *redis.Client, key interface{}, field interface{}, value interface{}) error {
	keyStr := fmt.Sprintf("%v", key)
	fieldStr := fmt.Sprintf("%v", field)
	jsonValue, err := rc.HGet(ctx, keyStr, fieldStr).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonValue), value)
}

// list
func SetListRedis(ctx context.Context, rc *redis.Client, key interface{}, value interface{}, expiration time.Duration) error {
	keyStr := fmt.Sprintf("%v", key)
	if valueStr, ok := value.(string); ok {
		return rc.LPush(ctx, keyStr, valueStr).Err()
	}
	// 序列化值为 JSON 字符串
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rc.LPush(ctx, keyStr, jsonValue).Err()
}

func GetListRedis(ctx context.Context, rc *redis.Client, key interface{}, value interface{}) error {
	keyStr := fmt.Sprintf("%v", key)
	jsonValue, err := rc.LRange(ctx, keyStr, 0, -1).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonValue[0]), value)
}

// 分布式锁相关方法

// AcquireDistributedLock 获取分布式锁
func AcquireDistributedLock(ctx context.Context, rc *redis.Client, key interface{}, expiration time.Duration) bool {
	keyStr := fmt.Sprintf("%v", key)
	// 使用Redis SETNX命令获取锁
	ok, err := rc.SetNX(ctx, keyStr, "1", expiration).Result()
	if err != nil {
		fmt.Printf("Failed to acquire distributed lock for %s: %v\n", keyStr, err)
		return false
	}
	return ok
}

// ReleaseDistributedLock 释放分布式锁
func ReleaseDistributedLock(ctx context.Context, rc *redis.Client, key interface{}) error {
	keyStr := fmt.Sprintf("%v", key)
	// 删除锁
	return rc.Del(ctx, keyStr).Err()
}
