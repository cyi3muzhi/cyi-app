package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// AdminCategory 分类领域模型
type AdminCategory struct {
	ID          int64
	ParentID    int64
	Name        string
	Slug        string
	Icon        string
	Image       string
	Sort        int32
	Status      bool
	Showcase    bool
	Count       int32
	Level       int32
	SeoTitle    string
	SeoKeywords string
	SeoDesc     string
	CreatedAt   int64
	Children    []*AdminCategory
}

// CategoryRepo 分类数据仓库接口
type CategoryRepo interface {
	// ListCategoryTree 获取分类树
	ListCategoryTree(ctx context.Context) ([]*AdminCategory, int32, error)
	// ListChildCategories 获取子分类列表
	ListChildCategories(ctx context.Context, parentID int64) ([]*AdminCategory, int32, error)
	// GetCategory 获取单个分类
	GetCategory(ctx context.Context, id int64) (*AdminCategory, error)
	// CreateCategory 创建分类
	CreateCategory(ctx context.Context, in *AdminCategory) (int64, error)
	// UpdateCategory 更新分类
	UpdateCategory(ctx context.Context, in *AdminCategory) error
	// DeleteCategory 删除分类
	DeleteCategory(ctx context.Context, id int64) error
	// UpdateCategorySort 更新排序
	UpdateCategorySort(ctx context.Context, id int64, sort int32) error
	// UpdateCategoryStatus 更新状态
	UpdateCategoryStatus(ctx context.Context, id int64, status bool) error
	// ToggleCategoryShowcase 切换首页展示
	ToggleCategoryShowcase(ctx context.Context, id int64, showcase bool) error
	// BatchCategoryOperation 批量操作分类
	BatchCategoryOperation(ctx context.Context, action string, ids []int64) (int32, error)
}

// CategoryUseCase 分类管理业务用例
type CategoryUseCase struct {
	log  *log.Helper
	repo CategoryRepo
}

// NewCategoryUseCase 创建分类管理业务用例
func NewCategoryUseCase(repo CategoryRepo, logger log.Logger) *CategoryUseCase {
	return &CategoryUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

// ListCategoryTree 获取分类树
func (uc *CategoryUseCase) ListCategoryTree(ctx context.Context) ([]*AdminCategory, int32, error) {
	return uc.repo.ListCategoryTree(ctx)
}

// ListChildCategories 获取子分类列表
func (uc *CategoryUseCase) ListChildCategories(ctx context.Context, parentID int64) ([]*AdminCategory, int32, error) {
	return uc.repo.ListChildCategories(ctx, parentID)
}

// GetCategory 获取单个分类
func (uc *CategoryUseCase) GetCategory(ctx context.Context, id int64) (*AdminCategory, error) {
	return uc.repo.GetCategory(ctx, id)
}

// CreateCategory 创建分类
func (uc *CategoryUseCase) CreateCategory(ctx context.Context, in *AdminCategory) (int64, error) {
	return uc.repo.CreateCategory(ctx, in)
}

// UpdateCategory 更新分类
func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, in *AdminCategory) error {
	return uc.repo.UpdateCategory(ctx, in)
}

// DeleteCategory 删除分类
func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, id int64) error {
	return uc.repo.DeleteCategory(ctx, id)
}

// UpdateCategorySort 更新排序
func (uc *CategoryUseCase) UpdateCategorySort(ctx context.Context, id int64, sort int32) error {
	return uc.repo.UpdateCategorySort(ctx, id, sort)
}

// UpdateCategoryStatus 更新状态
func (uc *CategoryUseCase) UpdateCategoryStatus(ctx context.Context, id int64, status bool) error {
	return uc.repo.UpdateCategoryStatus(ctx, id, status)
}

// ToggleCategoryShowcase 切换首页展示
func (uc *CategoryUseCase) ToggleCategoryShowcase(ctx context.Context, id int64, showcase bool) error {
	return uc.repo.ToggleCategoryShowcase(ctx, id, showcase)
}

// BatchCategoryOperation 批量操作分类
func (uc *CategoryUseCase) BatchCategoryOperation(ctx context.Context, action string, ids []int64) (int32, error) {
	return uc.repo.BatchCategoryOperation(ctx, action, ids)
}
