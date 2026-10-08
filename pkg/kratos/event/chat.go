package event

const (
	// SubjectChatMessage 聊天消息推送主题
	// Chat Service 发送消息落库后发布，interface/admin 网关各自 Subscribe 订阅，
	// 按 receiver_id 匹配本地 Hub 连接实现实时推送（无队列组，靠连接本地存在性保证不重复）。
	SubjectChatMessage = "chat.message"

	// SubjectChatConversation 会话状态变更主题（新建/接入/转接/结束），M3 客服工作台实时刷新使用
	SubjectChatConversation = "chat.conversation"
)

// 接收方类型
const (
	ChatReceiverUser  = int32(1) // 用户
	ChatReceiverAgent = int32(2) // 客服
)

// 会话状态事件类型
const (
	ChatEventCreated     = int32(1) // 新建
	ChatEventAccepted    = int32(2) // 客服接入
	ChatEventTransferred = int32(3) // 转接
	ChatEventClosed      = int32(4) // 结束
)

// ChatMessageEvent 聊天消息推送事件
// Chat Service 发送消息统一落库后发布，网关按接收方类型路由推送。
type ChatMessageEvent struct {
	ConversationID int64  `json:"conversation_id"` // 会话ID
	MessageID      int64  `json:"message_id"`      // 消息ID
	SenderType     int32  `json:"sender_type"`     // 发送者类型 1=用户 2=客服 3=系统 4=AI
	SenderID       int64  `json:"sender_id"`       // 发送者ID
	ReceiverType   int32  `json:"receiver_type"`   // 接收方类型 1=用户 2=客服
	ReceiverID     int64  `json:"receiver_id"`     // 接收方ID
	MsgType        int32  `json:"msg_type"`        // 消息类型
	Content        string `json:"content"`         // 内容
	Extra          string `json:"extra"`           // 扩展JSON
	CreatedAt      int64  `json:"created_at"`      // 发送时间
}

// ChatConversationEvent 会话状态事件（M3 使用）
type ChatConversationEvent struct {
	ConversationID int64 `json:"conversation_id"` // 会话ID
	EventType      int32 `json:"event_type"`      // 1=新建 2=接入 3=转接 4=结束
	UserID         int64 `json:"user_id"`         // 用户ID
	AgentID        int64 `json:"agent_id"`        // 客服ID
}
