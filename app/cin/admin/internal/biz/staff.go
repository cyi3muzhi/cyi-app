package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Staff 员工实体
type Staff struct {
	ID         int64
	Username   string
	Name       string
	Role       string
	Email      string
	Phone      string
	Department string
	Position   string
	Bio        string
	Avatar     string
	Status     int
	LastLogin  string
	CreatedAt  string
}

// StaffRepo 员工仓库接口
type StaffRepo interface {
	List(ctx context.Context, page, pageSize int32, keyword, status string) ([]*Staff, int32, error)
	Create(ctx context.Context, s *Staff, password string) (int64, error)
	Update(ctx context.Context, s *Staff) error
	ToggleStatus(ctx context.Context, id int64) (int, error)
}

// StaffUseCase 员工管理业务逻辑
type StaffUseCase struct {
	repo StaffRepo
	log  *log.Helper
}

// NewStaffUseCase 创建 StaffUseCase
func NewStaffUseCase(repo StaffRepo, logger log.Logger) *StaffUseCase {
	return &StaffUseCase{repo: repo, log: log.NewHelper(logger)}
}

// List 获取员工列表（带分页）
func (uc *StaffUseCase) List(ctx context.Context, page, pageSize int32, keyword, status string) ([]*Staff, int32, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return uc.repo.List(ctx, page, pageSize, keyword, status)
}

// Create 创建员工
func (uc *StaffUseCase) Create(ctx context.Context, s *Staff, password string) (int64, error) {
	return uc.repo.Create(ctx, s, password)
}

// Update 更新员工信息
func (uc *StaffUseCase) Update(ctx context.Context, s *Staff) error {
	return uc.repo.Update(ctx, s)
}

// ToggleStatus 切换员工状态
func (uc *StaffUseCase) ToggleStatus(ctx context.Context, id int64) (int, error) {
	return uc.repo.ToggleStatus(ctx, id)
}
