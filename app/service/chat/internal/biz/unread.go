package biz

import (
	"context"
	"fmt"

	"cyi-app/pkg/kratos/event"

	"github.com/redis/go-redis/v9"
)

// 未读数 Redis key：chat:unread:conv:{conversationID}:{user|agent}
// 字符串 INCR 计数；side 表示"哪一方的未读数"：
//   - user：发给用户的消息未读数（客服/AI 发的）
//   - agent：发给客服的消息未读数（用户发的）
const unreadConvKeyPrefix = "chat:unread:conv:"

func unreadKey(conversationID int64, side string) string {
	return fmt.Sprintf("%s%d:%s", unreadConvKeyPrefix, conversationID, side)
}

func unreadSide(receiverType int32) string {
	if receiverType == event.ChatReceiverAgent {
		return "agent"
	}
	return "user"
}

// IncrUnread 对接收方未读数 +1
func IncrUnread(ctx context.Context, rdb redis.UniversalClient, conversationID int64, receiverType int32) {
	if rdb == nil || conversationID <= 0 {
		return
	}
	rdb.Incr(ctx, unreadKey(conversationID, unreadSide(receiverType)))
}

// ClearUnread 清空某一方的未读数（标记已读时调用）
func ClearUnread(ctx context.Context, rdb redis.UniversalClient, conversationID int64, receiverType int32) {
	if rdb == nil || conversationID <= 0 {
		return
	}
	rdb.Del(ctx, unreadKey(conversationID, unreadSide(receiverType)))
}

// GetUnread 读取某一方未读数；ok=false 表示 Redis 无记录（可回退 DB 统计）
func GetUnread(ctx context.Context, rdb redis.UniversalClient, conversationID int64, receiverType int32) (int32, bool) {
	if rdb == nil || conversationID <= 0 {
		return 0, false
	}
	v, err := rdb.Get(ctx, unreadKey(conversationID, unreadSide(receiverType))).Int64()
	if err != nil {
		return 0, false
	}
	return int32(v), true
}

// DelConvUnread 删除会话两侧未读数（会话结束时清理）
func DelConvUnread(ctx context.Context, rdb redis.UniversalClient, conversationID int64) {
	if rdb == nil || conversationID <= 0 {
		return
	}
	rdb.Del(ctx, unreadKey(conversationID, "user"), unreadKey(conversationID, "agent"))
}
