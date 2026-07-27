package service

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
)

func (s *CinInterface) ListCategoryTree(ctx context.Context, req *v1.ListCategoryTreeRequest) (*v1.ListCategoryTreeReply, error) {
	nodes, err := s.catUC.ListCategoryTree(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.CategoryNode, 0, len(nodes))
	for _, n := range nodes {
		items = append(items, convertCategoryNode(n))
	}
	return &v1.ListCategoryTreeReply{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

func (s *CinInterface) GetCategory(ctx context.Context, req *v1.GetCategoryRequest) (*v1.GetCategoryReply, error) {
	node, err := s.catUC.GetCategory(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetCategoryReply{
		Item: convertCategoryNode(node),
	}, nil
}

func convertCategoryNode(n *biz.CategoryNode) *v1.CategoryNode {
	if n == nil {
		return nil
	}
	children := make([]*v1.CategoryNode, 0, len(n.Children))
	for _, c := range n.Children {
		children = append(children, convertCategoryNode(c))
	}
	return &v1.CategoryNode{
		Id:       n.ID,
		Name:     n.Name,
		Slug:     n.Slug,
		Icon:     n.Icon,
		Sort:     n.Sort,
		Status:   n.Status,
		Showcase: n.Showcase,
		Count:    n.Count,
		Image:    n.Image,
		ParentId: n.ParentID,
		Children: children,
	}
}
