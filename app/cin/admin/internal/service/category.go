package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

// ListCategoryTree 获取分类树
func (s *AdminService) ListCategoryTree(ctx context.Context, req *adminv1.ListCategoryTreeRequest) (*adminv1.ListCategoryTreeReply, error) {
	items, total, err := s.cc.ListCategoryTree(ctx)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*adminv1.CategoryTreeItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, bizCategoryToProto(item))
	}
	return &adminv1.ListCategoryTreeReply{
		Items: replyItems,
		Total: total,
	}, nil
}

// ListChildCategories 获取子分类列表
func (s *AdminService) ListChildCategories(ctx context.Context, req *adminv1.ListChildCategoriesRequest) (*adminv1.ListChildCategoriesReply, error) {
	items, total, err := s.cc.ListChildCategories(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*adminv1.CategoryTreeItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, bizCategoryToProto(item))
	}
	return &adminv1.ListChildCategoriesReply{
		Items: replyItems,
		Total: total,
	}, nil
}

// GetCategory 获取单个分类
func (s *AdminService) GetCategory(ctx context.Context, req *adminv1.GetCategoryRequest) (*adminv1.GetCategoryReply, error) {
	item, err := s.cc.GetCategory(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &adminv1.GetCategoryReply{
		Item: bizCategoryToProto(item),
	}, nil
}

// CreateCategory 创建分类
func (s *AdminService) CreateCategory(ctx context.Context, req *adminv1.CreateCategoryRequest) (*adminv1.CreateCategoryReply, error) {
	id, err := s.cc.CreateCategory(ctx, &biz.AdminCategory{
		ParentID:    req.ParentId,
		Name:        req.Name,
		Slug:        req.Slug,
		Icon:        req.Icon,
		Sort:        req.Sort,
		Status:      req.Status,
		Showcase:    req.Showcase,
		Image:       req.Image,
		SeoTitle:    req.SeoTitle,
		SeoKeywords: req.SeoKeywords,
		SeoDesc:     req.SeoDesc,
	})
	if err != nil {
		return nil, err
	}
	return &adminv1.CreateCategoryReply{
		Id: id,
	}, nil
}

// UpdateCategory 更新分类
func (s *AdminService) UpdateCategory(ctx context.Context, req *adminv1.UpdateCategoryRequest) (*adminv1.UpdateCategoryReply, error) {
	err := s.cc.UpdateCategory(ctx, &biz.AdminCategory{
		ID:          req.Id,
		ParentID:    req.ParentId,
		Name:        req.Name,
		Slug:        req.Slug,
		Icon:        req.Icon,
		Sort:        req.Sort,
		Status:      req.Status,
		Showcase:    req.Showcase,
		Image:       req.Image,
		SeoTitle:    req.SeoTitle,
		SeoKeywords: req.SeoKeywords,
		SeoDesc:     req.SeoDesc,
	})
	if err != nil {
		return nil, err
	}
	return &adminv1.UpdateCategoryReply{
		Success: true,
	}, nil
}

// DeleteCategory 删除分类
func (s *AdminService) DeleteCategory(ctx context.Context, req *adminv1.DeleteCategoryRequest) (*adminv1.DeleteCategoryReply, error) {
	err := s.cc.DeleteCategory(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &adminv1.DeleteCategoryReply{
		Success: true,
	}, nil
}

// UpdateCategorySort 更新排序
func (s *AdminService) UpdateCategorySort(ctx context.Context, req *adminv1.UpdateCategorySortRequest) (*adminv1.UpdateCategorySortReply, error) {
	err := s.cc.UpdateCategorySort(ctx, req.Id, req.Sort)
	if err != nil {
		return nil, err
	}
	return &adminv1.UpdateCategorySortReply{
		Success: true,
	}, nil
}

// UpdateCategoryStatus 更新状态
func (s *AdminService) UpdateCategoryStatus(ctx context.Context, req *adminv1.UpdateCategoryStatusRequest) (*adminv1.UpdateCategoryStatusReply, error) {
	err := s.cc.UpdateCategoryStatus(ctx, req.Id, req.Status)
	if err != nil {
		return nil, err
	}
	return &adminv1.UpdateCategoryStatusReply{
		Success: true,
	}, nil
}

// ToggleCategoryShowcase 切换首页展示
func (s *AdminService) ToggleCategoryShowcase(ctx context.Context, req *adminv1.ToggleCategoryShowcaseRequest) (*adminv1.ToggleCategoryShowcaseReply, error) {
	err := s.cc.ToggleCategoryShowcase(ctx, req.Id, req.Showcase)
	if err != nil {
		return nil, err
	}
	return &adminv1.ToggleCategoryShowcaseReply{
		Success: true,
	}, nil
}

// BatchCategoryOperation 批量操作分类
func (s *AdminService) BatchCategoryOperation(ctx context.Context, req *adminv1.BatchCategoryOperationRequest) (*adminv1.BatchCategoryOperationReply, error) {
	affected, err := s.cc.BatchCategoryOperation(ctx, req.Action, req.Ids)
	if err != nil {
		return nil, err
	}
	return &adminv1.BatchCategoryOperationReply{
		Success:       true,
		AffectedCount: affected,
	}, nil
}

// bizCategoryToProto 将 biz 领域模型转换为 proto 消息
func bizCategoryToProto(item *biz.AdminCategory) *adminv1.CategoryTreeItem {
	if item == nil {
		return nil
	}
	pb := &adminv1.CategoryTreeItem{
		Id:          item.ID,
		ParentId:    item.ParentID,
		Name:        item.Name,
		Slug:        item.Slug,
		Icon:        item.Icon,
		Image:       item.Image,
		Sort:        item.Sort,
		Status:      item.Status,
		Showcase:    item.Showcase,
		Count:       item.Count,
		SeoTitle:    item.SeoTitle,
		SeoKeywords: item.SeoKeywords,
		SeoDesc:     item.SeoDesc,
		CreatedAt:   item.CreatedAt,
	}
	if len(item.Children) > 0 {
		pb.Children = make([]*adminv1.CategoryTreeItem, 0, len(item.Children))
		for _, child := range item.Children {
			pb.Children = append(pb.Children, bizCategoryToProto(child))
		}
	}
	return pb
}
