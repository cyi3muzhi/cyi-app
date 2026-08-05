package models

import (
	"encoding/json"

	"cyi-app/pkg/util/gorm"
)

// Message 消息（管理后台发送的通知/公告）
type Message struct {
	gorm.Model
	Title         string          `gorm:"size:128;not null;comment:消息标题"`            // 标题
	Content       string          `gorm:"type:text;not null;comment:消息内容"`          // 内容（支持换行文本）
	MsgType       int32           `gorm:"not null;index;comment:消息类型"`              // 消息类型 1=订单 2=系统 3=营销
	TargetType    int32           `gorm:"not null;comment:发送范围"`                    // 发送范围 0=全体用户 1=指定用户
	TargetUserIDs json.RawMessage `gorm:"type:json;comment:指定用户ID列表(JSON数组)"`      // 指定用户ID列表（仅 target_type=1 时有值）
	Status        int32           `gorm:"not null;default:1;comment:消息状态"`          // 消息状态 1=启用 2=下线
	CreatedBy     int64           `gorm:"not null;comment:创建人管理员ID"`               // 创建人（管理员ID）
}

// TableName 指定表名
func (Message) TableName() string {
	return "message"
}

// UserMessage 用户消息状态（已读标记，按用户维度）
type UserMessage struct {
	gorm.Model
	UserID    int64 `gorm:"not null;index:idx_user_message,unique,priority:1;comment:用户ID"` // 用户ID
	MessageID int64 `gorm:"not null;index:idx_user_message,unique,priority:2;comment:消息ID"` // 消息ID
	IsRead    int32 `gorm:"not null;default:0;comment:是否已读"`                               // 是否已读 0=未读 1=已读
	ReadAt    int64 `gorm:"default:0;comment:已读时间"`                                        // 已读时间（Unix 秒）
}

// TableName 指定表名
func (UserMessage) TableName() string {
	return "user_message"
}
