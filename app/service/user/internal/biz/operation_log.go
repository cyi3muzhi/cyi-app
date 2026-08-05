package biz

import (
	"context"
)

// OperationLogItem 操作日志查询项
type OperationLogItem struct {
	ID        int64
	Operator  string // 操作人姓名
	Role      string // 角色标识（如 super_admin）
	Module    string // 操作模块
	Action    string // 动作
	Content   string // 操作内容
	Method    string // HTTP 方法
	Path      string // 请求路径
	IP        string // 来源 IP
	UserAgent string // User-Agent
	Result    int32  // 执行结果：1-成功 0-失败
	CreatedAt int64  // 操作时间（Unix 秒）
	Avatar    string // 操作人头像（OSS 路径，由 admin 服务转公开 URL）
}

// OperationLogStats 操作日志统计
type OperationLogStats struct {
	Today int64 // 今日操作数（今日零点起）
	Week  int64 // 近7天操作数（滚动窗口）
	Month int64 // 近30天操作数（滚动窗口）
	Total int64 // 累计操作数
}

// OperationLogRecord 操作日志写入实体（来自 admin 服务 NATS 事件）
type OperationLogRecord struct {
	AdminUserID  int64
	OperatorName string
	RoleName     string
	Module       string
	Action       string
	Content      string
	Method       string
	Path         string
	IP           string
	UserAgent    string
	RequestBody  string
	Result       int32
	StatusCode   int32
	DurationMs   int64
	CreatedAt    int64
}

// OperationLogRepo 操作日志仓储接口
type OperationLogRepo interface {
	// Create 写入一条操作日志（NATS 订阅消费时调用）
	Create(ctx context.Context, in *OperationLogRecord) error
	// List 分页查询操作日志（支持日期/操作人/模块/关键词筛选）
	List(ctx context.Context, page, pageSize int32, startDate, endDate, operator, module, keyword string) ([]*OperationLogItem, int32, error)
	// Count 统计操作日志数量（今日/本周/本月/累计）
	Count(ctx context.Context) (*OperationLogStats, error)
	// ListOperators 获取操作人列表（按姓名去重）
	ListOperators(ctx context.Context) ([]string, error)
}

// OperationLogUseCase 操作日志用例
type OperationLogUseCase struct {
	repo OperationLogRepo
}

// NewOperationLogUseCase 创建操作日志用例
func NewOperationLogUseCase(repo OperationLogRepo) *OperationLogUseCase {
	return &OperationLogUseCase{repo: repo}
}

// Create 写入操作日志
func (uc *OperationLogUseCase) Create(ctx context.Context, in *OperationLogRecord) error {
	return uc.repo.Create(ctx, in)
}

// List 分页查询操作日志
func (uc *OperationLogUseCase) List(ctx context.Context, page, pageSize int32, startDate, endDate, operator, module, keyword string) ([]*OperationLogItem, int32, error) {
	return uc.repo.List(ctx, page, pageSize, startDate, endDate, operator, module, keyword)
}

// Count 统计操作日志数量
func (uc *OperationLogUseCase) Count(ctx context.Context) (*OperationLogStats, error) {
	return uc.repo.Count(ctx)
}

// ListOperators 获取操作人列表
func (uc *OperationLogUseCase) ListOperators(ctx context.Context) ([]string, error) {
	return uc.repo.ListOperators(ctx)
}
