package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// CategoryRepo 商品分类仓储接口
type CategoryRepo interface {
	// ListCategoryTree 获取分类树（返回全量扁平列表）
	ListCategoryTree(ctx context.Context, parentID int64) ([]*Category, error)
	// GetCategory 获取单个分类
	GetCategory(ctx context.Context, id int64) (*Category, error)
	// CreateCategory 创建分类
	CreateCategory(ctx context.Context, in *Category) error
	// UpdateCategory 更新分类
	UpdateCategory(ctx context.Context, in *Category) error
	// DeleteCategory 删除分类
	DeleteCategory(ctx context.Context, id int64) error
	// ListChildCategories 获取直接子分类列表
	ListChildCategories(ctx context.Context, parentID int64) ([]*Category, error)
	// UpdateCategorySort 更新分类排序
	UpdateCategorySort(ctx context.Context, id int64, sort int32) error
	// UpdateCategoryStatus 更新分类状态
	UpdateCategoryStatus(ctx context.Context, id int64, status bool) error
	// ToggleCategoryShowcase 切换首页展示
	ToggleCategoryShowcase(ctx context.Context, id int64, showcase bool) error
	// BatchOperation 批量操作（action: delete/disable/enable）
	BatchOperation(ctx context.Context, action string, ids []int64) (int32, error)
}

// CategoryUseCase 商品分类业务逻辑
type CategoryUseCase struct {
	repo CategoryRepo
	log  *log.Helper
}

// NewCategoryUseCase 创建分类 UseCase
func NewCategoryUseCase(repo CategoryRepo, logger log.Logger) *CategoryUseCase {
	return &CategoryUseCase{repo: repo, log: log.NewHelper(logger)}
}

// ListCategoryTree 获取分类树，按 parentID 过滤并组装树形结构
func (uc *CategoryUseCase) ListCategoryTree(ctx context.Context, parentID int64) ([]*Category, error) {
	// 从 repo 获取扁平列表
	all, err := uc.repo.ListCategoryTree(ctx, parentID)
	if err != nil {
		return nil, err
	}
	// 按 Sort 排序后组装树结构
	uc.sortBySort(all)
	tree := uc.buildTree(all, parentID)
	return tree, nil
}

// GetCategory 获取单个分类
func (uc *CategoryUseCase) GetCategory(ctx context.Context, id int64) (*Category, error) {
	return uc.repo.GetCategory(ctx, id)
}

// CreateCategory 创建分类
func (uc *CategoryUseCase) CreateCategory(ctx context.Context, in *Category) error {
	return uc.repo.CreateCategory(ctx, in)
}

// UpdateCategory 更新分类
func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, in *Category) error {
	return uc.repo.UpdateCategory(ctx, in)
}

// DeleteCategory 删除分类
func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, id int64) error {
	return uc.repo.DeleteCategory(ctx, id)
}

// ListChildCategories 获取直接子分类
func (uc *CategoryUseCase) ListChildCategories(ctx context.Context, parentID int64) ([]*Category, error) {
	return uc.repo.ListChildCategories(ctx, parentID)
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

// BatchOperation 批量操作
func (uc *CategoryUseCase) BatchOperation(ctx context.Context, action string, ids []int64) (int32, error) {
	return uc.repo.BatchOperation(ctx, action, ids)
}

// sortBySort 按 Sort 字段对分类列表排序（简单插入排序，保持稳定）
func (uc *CategoryUseCase) sortBySort(list []*Category) {
	for i := 1; i < len(list); i++ {
		key := list[i]
		j := i - 1
		for j >= 0 && list[j].Sort > key.Sort {
			list[j+1] = list[j]
			j--
		}
		list[j+1] = key
	}
}

// buildTree 将扁平分类列表组装为树形结构
// 当 parentID > 0 时，只返回该父节点下的子树
func (uc *CategoryUseCase) buildTree(all []*Category, parentID int64) []*Category {
	// 按 parentID 分组
	childrenMap := make(map[int64][]*Category)
	for _, cat := range all {
		pid := cat.ParentID
		childrenMap[pid] = append(childrenMap[pid], cat)
	}
	// 递归组装
	var root []*Category
	for _, cat := range all {
		if parentID > 0 {
			// 只处理指定父节点下的子分类
			if cat.ParentID == parentID {
				uc.fillChildren(cat, childrenMap)
				root = append(root, cat)
			}
		} else {
			// 返回完整树：顶级分类（ParentID == 0）
			if cat.ParentID == 0 {
				uc.fillChildren(cat, childrenMap)
				root = append(root, cat)
			}
		}
	}
	return root
}

// fillChildren 递归填充子分类
func (uc *CategoryUseCase) fillChildren(parent *Category, childrenMap map[int64][]*Category) {
	if children, ok := childrenMap[parent.ID]; ok {
		parent.Children = children
		for _, child := range children {
			uc.fillChildren(child, childrenMap)
		}
	}
}
