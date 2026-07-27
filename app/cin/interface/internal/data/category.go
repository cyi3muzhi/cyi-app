package data

import (
	"context"

	productV1 "cyi-app/api/service/product/v1"
	"cyi-app/app/cin/interface/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type categoryRepo struct {
	log  *log.Helper
	data *Data
}

func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &categoryRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

func (r *categoryRepo) ListCategoryTree(ctx context.Context, parentID int64) ([]*biz.CategoryNode, error) {
	resp, err := r.data.cc.ListCategoryTree(ctx, &productV1.ListCategoryTreeRequest{
		ParentId: parentID,
	})
	if err != nil {
		return nil, err
	}
	return convertCategoryNodes(resp.Items), nil
}

func (r *categoryRepo) GetCategory(ctx context.Context, id int64) (*biz.CategoryNode, error) {
	resp, err := r.data.cc.GetCategory(ctx, &productV1.GetCategoryRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return convertCategoryNode(resp.Item), nil
}

func convertCategoryNodes(nodes []*productV1.CategoryNode) []*biz.CategoryNode {
	result := make([]*biz.CategoryNode, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, convertCategoryNode(n))
	}
	return result
}

func convertCategoryNode(n *productV1.CategoryNode) *biz.CategoryNode {
	if n == nil {
		return nil
	}
	return &biz.CategoryNode{
		ID:        n.Id,
		Name:      n.Name,
		Slug:      n.Slug,
		Icon:      n.Icon,
		Sort:      n.Sort,
		Status:    n.Status,
		Showcase:  n.Showcase,
		Count:     n.Count,
		Image:     n.Image,
		ParentID:  n.ParentId,
		Children:  convertCategoryNodes(n.Children),
	}
}