package models

import "cyi-app/pkg/util/gorm"

// Conversation 会话
type Conversation struct {
	gorm.Model
	UserID        int64  `json:"user_id" gorm:"not null;index:idx_user_status,priority:1;comment:用户ID(发起方)"`        // 用户ID
	AgentID       int64  `json:"agent_id" gorm:"not null;default:0;index:idx_agent_status,priority:1;comment:客服ID"` // 客服ID（admin_user.id，0=未分配）
	Status        int32  `json:"status" gorm:"not null;default:1;index:idx_user_status,priority:2;comment:会话状态"`    // 1=待接入 2=进行中 3=已结束
	ServiceType   int32  `json:"service_type" gorm:"not null;default:1;comment:接待类型"`                               // 1=人工 2=AI接待中
	LastMsgType   int32  `json:"last_msg_type" gorm:"not null;default:1;comment:最后消息类型"`                            // 1=文本 2=图片 3=商品卡片 4=订单卡片 5=系统
	LastMessage   string `json:"last_message" gorm:"size:255;comment:最后消息摘要"`                                       // 最后一条消息摘要
	LastMessageAt int64  `json:"last_message_at" gorm:"not null;default:0;comment:最后消息时间"`                          // 最后一条消息时间（Unix 秒）
	EndedAt       int64  `json:"ended_at" gorm:"not null;default:0;comment:结束时间"`                                   // 结束时间（Unix 秒）
}

// TableName 指定表名
func (Conversation) TableName() string {
	return "conversation"
}

// IConversation gorm gen 接口（可扩展自定义查询）
type IConversation interface {
}
