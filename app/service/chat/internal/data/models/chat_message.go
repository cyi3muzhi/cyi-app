package models

import "cyi-app/pkg/util/gorm"

// ChatMessage 聊天消息
type ChatMessage struct {
	gorm.Model
	ConversationID int64  `json:"conversation_id" gorm:"not null;index:idx_conv_id,priority:1;comment:会话ID"`   // 会话ID
	SenderType     int32  `json:"sender_type" gorm:"not null;comment:发送者类型"`                                   // 1=用户 2=客服 3=系统 4=AI
	SenderID       int64  `json:"sender_id" gorm:"not null;comment:发送者ID"`                                     // 发送者ID
	MsgType        int32  `json:"msg_type" gorm:"not null;default:1;comment:消息类型"`                             // 1=文本 2=图片 3=商品卡片 4=订单卡片 5=系统
	Content        string `json:"content" gorm:"type:text;not null;comment:消息内容"`                              // 文本内容
	Extra          string `json:"extra" gorm:"type:text;comment:扩展数据(JSON)"`                                   // 扩展数据（图片URL/卡片JSON）
	IsRead         int32  `json:"is_read" gorm:"not null;default:0;index:idx_conv_id,priority:2;comment:是否已读"` // 0=未读 1=已读
	ReadAt         int64  `json:"read_at" gorm:"not null;default:0;comment:已读时间"`                              // 已读时间（Unix 秒）
}

// TableName 指定表名
func (ChatMessage) TableName() string {
	return "chat_message"
}

// IChatMessage gorm gen 接口（可扩展自定义查询）
type IChatMessage interface {
}
