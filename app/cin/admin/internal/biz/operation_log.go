package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// OperationLogItem 操作日志查询项
type OperationLogItem struct {
	ID        int64
	Operator  string // 操作人姓名
	Role      string // 角色标识
	Module    string // 操作模块
	Action    string // 动作
	Content   string // 操作内容
	Method    string // HTTP 方法
	Path      string // 请求路径
	IP        string // 来源 IP
	UserAgent string // User-Agent
	Result    int32  // 执行结果：1-成功 0-失败
	CreatedAt int64  // 操作时间（Unix 秒）
	Avatar    string // 操作人头像（公开 URL）
}

// OperationLogStats 操作日志统计
type OperationLogStats struct {
	Today int64
	Week  int64
	Month int64
	Total int64
}

// OperationLogRepo 操作日志数据仓库接口
type OperationLogRepo interface {
	// List 分页查询操作日志
	List(ctx context.Context, page, pageSize int32, startDate, endDate, operator, module, keyword string) ([]*OperationLogItem, int32, error)
	// Count 统计操作日志数量
	Count(ctx context.Context) (*OperationLogStats, error)
	// ListOperators 获取操作人列表
	ListOperators(ctx context.Context) ([]string, error)
}

// OperationLogUseCase 操作日志用例
type OperationLogUseCase struct {
	repo OperationLogRepo
	log  *log.Helper
}

// NewOperationLogUseCase 创建操作日志用例
func NewOperationLogUseCase(repo OperationLogRepo, logger log.Logger) *OperationLogUseCase {
	return &OperationLogUseCase{repo: repo, log: log.NewHelper(logger)}
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
