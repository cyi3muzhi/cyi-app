package event

const (
	// SubjectAdminOperation 管理端操作日志事件主题
	// admin 服务在 HTTP 中间件中发布，user 服务订阅后写入操作日志表
	SubjectAdminOperation = "admin.operation"
)

// AdminOperationLog 管理端操作日志事件
type AdminOperationLog struct {
	AdminUserID  int64  `json:"admin_user_id"`  // 操作人 ID
	OperatorName string `json:"operator_name"`  // 操作人姓名/用户名
	RoleName     string `json:"role_name"`      // 角色标识（如 super_admin）
	Module       string `json:"module"`         // 操作模块（product/order/user/...）
	Action       string `json:"action"`         // 动作（新增/编辑/删除/上架/...）
	Content      string `json:"content"`        // 操作内容描述
	Method       string `json:"method"`         // HTTP 方法
	Path         string `json:"path"`           // 请求路径
	IP           string `json:"ip"`             // 来源 IP
	UserAgent    string `json:"user_agent"`     // User-Agent
	RequestBody  string `json:"request_body"`   // 请求体（截断后）
	Result       int32  `json:"result"`         // 执行结果：1=成功 0=失败
	StatusCode   int32  `json:"status_code"`    // HTTP 状态码
	DurationMs   int64  `json:"duration_ms"`    // 耗时（毫秒）
	CreatedAt    int64  `json:"created_at"`     // 操作时间（Unix 秒）
}
