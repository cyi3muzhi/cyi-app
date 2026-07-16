package service

import (
	"context"

	v1 "cyi-app/api/service/product/v1"
	"cyi-app/app/service/product/internal/biz"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
)

// CategoryService 商品分类 gRPC 服务
type CategoryService struct {
	v1.UnimplementedCategoryServer
	uc  *biz.CategoryUseCase
	log log.Logger
}

// NewCategoryService 创建分类服务
func NewCategoryService(uc *biz.CategoryUseCase, logger log.Logger) *CategoryService {
	return &CategoryService{
		uc:  uc,
		log: logger,
	}
}

// ListCategoryTree 获取分类树
func (s *CategoryService) ListCategoryTree(ctx context.Context, req *v1.ListCategoryTreeRequest) (*v1.ListCategoryTreeReply, error) {
	tree, err := s.uc.ListCategoryTree(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.CategoryNode, 0, len(tree))
	for _, cat := range tree {
		items = append(items, s.categoryToNode(cat))
	}
	return &v1.ListCategoryTreeReply{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

// GetCategory 获取单个分类
func (s *CategoryService) GetCategory(ctx context.Context, req *v1.GetCategoryRequest) (*v1.GetCategoryReply, error) {
	cat, err := s.uc.GetCategory(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetCategoryReply{
		Item: s.categoryToNode(cat),
	}, nil
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(ctx context.Context, req *v1.CreateCategoryRequest) (*v1.CreateCategoryReply, error) {
	in := &biz.Category{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	if err := s.uc.CreateCategory(ctx, in); err != nil {
		return nil, err
	}
	return &v1.CreateCategoryReply{
		Id: in.ID,
	}, nil
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(ctx context.Context, req *v1.UpdateCategoryRequest) (*v1.UpdateCategoryReply, error) {
	in := &biz.Category{
		ID:          req.Id,
		ParentID:    req.ParentId,
		Name:        req.Name,
		Slug:        req.Slug,
		Icon:        req.Icon,
		Image:       req.Image,
		Sort:        req.Sort,
		Status:      req.Status,
		Showcase:    req.Showcase,
		SeoTitle:    req.SeoTitle,
		SeoKeywords: req.SeoKeywords,
		SeoDesc:     req.SeoDesc,
	}
	if err := s.uc.UpdateCategory(ctx, in); err != nil {
		return nil, err
	}
	return &v1.UpdateCategoryReply{
		Success: true,
	}, nil
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(ctx context.Context, req *v1.DeleteCategoryRequest) (*v1.DeleteCategoryReply, error) {
	if err := s.uc.DeleteCategory(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteCategoryReply{
		Success: true,
	}, nil
}

// UpdateCategorySort 更新分类排序
func (s *CategoryService) UpdateCategorySort(ctx context.Context, req *v1.UpdateCategorySortRequest) (*v1.UpdateCategorySortReply, error) {
	if err := s.uc.UpdateCategorySort(ctx, req.Id, req.Sort); err != nil {
		return nil, err
	}
	return &v1.UpdateCategorySortReply{
		Success: true,
	}, nil
}

// UpdateCategoryStatus 更新分类状态
func (s *CategoryService) UpdateCategoryStatus(ctx context.Context, req *v1.UpdateCategoryStatusRequest) (*v1.UpdateCategoryStatusReply, error) {
	if err := s.uc.UpdateCategoryStatus(ctx, req.Id, req.Status); err != nil {
		return nil, err
	}
	return &v1.UpdateCategoryStatusReply{
		Success: true,
	}, nil
}

// ToggleCategoryShowcase 切换首页展示
func (s *CategoryService) ToggleCategoryShowcase(ctx context.Context, req *v1.ToggleCategoryShowcaseRequest) (*v1.ToggleCategoryShowcaseReply, error) {
	if err := s.uc.ToggleCategoryShowcase(ctx, req.Id, req.Showcase); err != nil {
		return nil, err
	}
	return &v1.ToggleCategoryShowcaseReply{
		Success: true,
	}, nil
}

// categoryToNode 将 Biz Category 转换为 proto CategoryNode
func (s *CategoryService) categoryToNode(cat *biz.Category) *v1.CategoryNode {
	node := &v1.CategoryNode{
		Id:          cat.ID,
		ParentId:    cat.ParentID,
		Name:        cat.Name,
		Slug:        cat.Slug,
		Icon:        cat.Icon,
		Image:       cat.Image,
		Sort:        cat.Sort,
		Status:      cat.Status,
		Showcase:    cat.Showcase,
		Count:       cat.Count,
		SeoTitle:    cat.SeoTitle,
		SeoKeywords: cat.SeoKeywords,
		SeoDesc:     cat.SeoDesc,
	}
	// 递归转换子分类
	if len(cat.Children) > 0 {
		children := make([]*v1.CategoryNode, 0, len(cat.Children))
		for _, child := range cat.Children {
			children = append(children, s.categoryToNode(child))
		}
		node.Children = children
	}
	return node
}
