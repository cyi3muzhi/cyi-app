package data

import (
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

// NewNATSConn 创建 NATS 连接（单例，最多重试 5 次，失败降级为 nil 不 panic）
// 优先从环境变量 NATS_URL 读取地址（Docker 环境配置为 nats://nats:4222），默认 localhost。
// 用于订阅 chat.message 主题，实时推送聊天消息给用户连接。
func NewNATSConn() *nats.Conn {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	for i := 0; i < 5; i++ {
		nc, err := nats.Connect(natsURL)
		if err == nil {
			return nc
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}
