package models

import "cyi-app/pkg/util/gorm"

// AdminOperationLog 管理端操作日志模型
// 由 admin 服务通过 NATS 发布事件，本服务订阅后写入
type AdminOperationLog struct {
	gorm.Model
	// 操作人ID
	AdminUserID int64 `json:"admin_user_id" gorm:"index;not null;default:0;comment:操作人ID"`
	// 操作人姓名/用户名
	OperatorName string `json:"operator_name" gorm:"type:varchar(64);default:'';comment:操作人姓名"`
	// 角色标识（如 super_admin）
	RoleName string `json:"role_name" gorm:"type:varchar(32);default:'';comment:角色标识"`
	// 操作模块（product/order/user/coupon/...）
	Module string `json:"module" gorm:"type:varchar(32);default:'';index;comment:操作模块"`
	// 动作（新增/编辑/删除/上架...）
	Action string `json:"action" gorm:"type:varchar(32);default:'';comment:动作"`
	// 操作内容描述
	Content string `json:"content" gorm:"type:varchar(500);default:'';comment:操作内容"`
	// HTTP 方法
	Method string `json:"method" gorm:"type:varchar(10);default:'';comment:HTTP方法"`
	// 请求路径
	Path string `json:"path" gorm:"type:varchar(200);default:'';comment:请求路径"`
	// 来源 IP
	IP string `json:"ip" gorm:"type:varchar(64);default:'';comment:来源IP"`
	// User-Agent
	UserAgent string `json:"user_agent" gorm:"type:varchar(300);default:'';comment:User-Agent"`
	// 请求体（脱敏截断）
	RequestBody string `json:"request_body" gorm:"type:text;comment:请求体"`
	// 执行结果：1-成功 0-失败
	Result int32 `json:"result" gorm:"not null;default:1;comment:执行结果 1成功0失败"`
	// HTTP 状态码
	StatusCode int32 `json:"status_code" gorm:"not null;default:0;comment:HTTP状态码"`
	// 耗时（毫秒）
	DurationMs int64 `json:"duration_ms" gorm:"not null;default:0;comment:耗时毫秒"`
}

// IAdminOperationLog 操作日志仓储接口标记
type IAdminOperationLog interface{}
