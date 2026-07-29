package biz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/blades/tools"
	"github.com/redis/go-redis/v9"
)

// RegisterPreferenceTool 注册用户偏好保存工具
func RegisterPreferenceTool(r *ToolRegistry, rdb redis.UniversalClient) {
	t, _ := tools.NewFunc("save_user_preference", "当用户告诉你他们的偏好设置时（如昵称、称呼方式、喜好等），调用此工具保存。参数 key 为偏好名称，value 为偏好值",
		func(ctx context.Context, input struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}) (string, error) {
			if rdb == nil {
				return "", fmt.Errorf("存储服务不可用")
			}

			// 从 context 获取用户信息
			userID := GetUserID(ctx)
			source := GetSource(ctx)
			if userID == 0 || source == "" {
				return "", fmt.Errorf("无法获取用户身份信息")
			}

			key := preferenceKey(source, userID)

			// 读取现有偏好
			var prefs map[string]string
			data, err := rdb.Get(ctx, key).Bytes()
			if err == nil {
				json.Unmarshal(data, &prefs)
			}
			if prefs == nil {
				prefs = make(map[string]string)
			}

			// 保存新偏好
			prefs[input.Key] = input.Value
			payload, _ := json.Marshal(prefs)
			if err := rdb.Set(ctx, key, payload, 0).Err(); err != nil {
				return "", fmt.Errorf("保存偏好失败: %w", err)
			}

			return fmt.Sprintf("已保存偏好设置：%s=%s", input.Key, input.Value), nil
		})
	r.Add(t)
}
