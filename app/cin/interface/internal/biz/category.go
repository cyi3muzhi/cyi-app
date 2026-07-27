package biz

import "context"

type CategoryNode struct {
	ID        int64
	Name      string
	Slug      string
	Icon      string
	Sort      int32
	Status    bool
	Showcase  bool
	Count     int32
	Image     string
	ParentID  int64
	Children  []*CategoryNode
}

type CategoryRepo interface {
	ListCategoryTree(ctx context.Context, parentID int64) ([]*CategoryNode, error)
	GetCategory(ctx context.Context, id int64) (*CategoryNode, error)
}

type CategoryUseCase struct {
	repo CategoryRepo
}

func NewCategoryUseCase(repo CategoryRepo) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

func (uc *CategoryUseCase) ListCategoryTree(ctx context.Context, parentID int64) ([]*CategoryNode, error) {
	return uc.repo.ListCategoryTree(ctx, parentID)
}

func (uc *CategoryUseCase) GetCategory(ctx context.Context, id int64) (*CategoryNode, error) {
	return uc.repo.GetCategory(ctx, id)
}