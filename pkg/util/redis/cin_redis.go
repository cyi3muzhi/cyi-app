package cin_redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
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

