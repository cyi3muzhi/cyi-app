package biz

// 会话状态
const (
	ConversationStatusPending = int32(1) // 待接入
	ConversationStatusActive  = int32(2) // 进行中
	ConversationStatusClosed  = int32(3) // 已结束
)

// 接待类型
const (
	ServiceTypeHuman = int32(1) // 人工接待
	ServiceTypeAI    = int32(2) // AI 接待中
)

// 发送者类型
const (
	SenderTypeUser   = int32(1) // 用户
	SenderTypeAgent  = int32(2) // 客服
	SenderTypeSystem = int32(3) // 系统
	SenderTypeAI     = int32(4) // AI
)

// 消息类型
const (
	MsgTypeText    = int32(1) // 文本
	MsgTypeImage   = int32(2) // 图片
	MsgTypeProduct = int32(3) // 商品卡片
	MsgTypeOrder   = int32(4) // 订单卡片
	MsgTypeSystem  = int32(5) // 系统提示
)

// Conversation 会话业务对象
type Conversation struct {
	ID            int64
	UserID        int64
	AgentID       int64
	Status        int32
	ServiceType   int32
	LastMsgType   int32
	LastMessage   string
	LastMessageAt int64
	EndedAt       int64
	CreateTime    int64
	UpdateTime    int64
	UnreadCount   int32 // 视图字段（按请求方视角，非表字段）
}

// ChatMessage 聊天消息业务对象
type ChatMessage struct {
	ID             int64
	ConversationID int64
	SenderType     int32
	SenderID       int64
	MsgType        int32
	Content        string
	Extra          string
	IsRead         int32
	ReadAt         int64
	CreateTime     int64
}
