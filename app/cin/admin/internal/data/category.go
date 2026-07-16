package data

import (
	"context"

	v1 "cyi-app/api/service/product/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type categoryRepo struct {
	data *Data
	log  *log.Helper
}

// NewCategoryRepo 创建分类数据仓库，通过 gRPC 调用商品服务
func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &categoryRepo{data: data, log: log.NewHelper(logger)}
}

// protoCategoryToBiz 将 proto CategoryNode 转为 biz AdminCategory
func protoCategoryToBiz(node *v1.CategoryNode) *biz.AdminCategory {
	if node == nil {
		return nil
	}
	c := &biz.AdminCategory{
		ID:          node.Id,
		ParentID:    node.ParentId,
		Name:        node.Name,
		Slug:        node.Slug,
		Icon:        node.Icon,
		Sort:        node.Sort,
		Status:      node.Status,
		Showcase:    node.Showcase,
		Count:       node.Count,
		Image:       node.Image,
		SeoTitle:    node.SeoTitle,
		SeoKeywords: node.SeoKeywords,
		SeoDesc:     node.SeoDesc,
	}
	if len(node.Children) > 0 {
		c.Children = make([]*biz.AdminCategory, 0, len(node.Children))
		for _, child := range node.Children {
			c.Children = append(c.Children, protoCategoryToBiz(child))
		}
	}
	return c
}

// bizCategoryToProto 将 biz AdminCategory 转为 proto CategoryNode
func bizCategoryToProto(c *biz.AdminCategory) *v1.CategoryNode {
	if c == nil {
		return nil
	}
	node := &v1.CategoryNode{
		Id:          c.ID,
		ParentId:    c.ParentID,
		Name:        c.Name,
		Slug:        c.Slug,
		Icon:        c.Icon,
		Sort:        c.Sort,
		Status:      c.Status,
		Showcase:    c.Showcase,
		Count:       c.Count,
		Image:       c.Image,
		SeoTitle:    c.SeoTitle,
		SeoKeywords: c.SeoKeywords,
		SeoDesc:     c.SeoDesc,
	}
	if len(c.Children) > 0 {
		node.Children = make([]*v1.CategoryNode, 0, len(c.Children))
		for _, child := range c.Children {
			node.Children = append(node.Children, bizCategoryToProto(child))
		}
	}
	return node
}

// ListCategoryTree 获取分类树
func (r *categoryRepo) ListCategoryTree(ctx context.Context) ([]*biz.AdminCategory, int32, error) {
	reply, err := r.data.cc.ListCategoryTree(ctx, &v1.ListCategoryTreeRequest{})
	if err != nil {
		return nil, 0, err
	}
	items := make([]*biz.AdminCategory, 0, len(reply.Items))
	for _, item := range reply.Items {
		items = append(items, protoCategoryToBiz(item))
	}
	return items, reply.Total, nil
}

// ListChildCategories 获取子分类列表
func (r *categoryRepo) ListChildCategories(ctx context.Context, parentID int64) ([]*biz.AdminCategory, int32, error) {
	reply, err := r.data.cc.ListCategoryTree(ctx, &v1.ListCategoryTreeRequest{ParentId: parentID})
	if err != nil {
		return nil, 0, err
	}
	items := make([]*biz.AdminCategory, 0, len(reply.Items))
	for _, item := range reply.Items {
		items = append(items, protoCategoryToBiz(item))
	}
	return items, reply.Total, nil
}

// GetCategory 获取单个分类
func (r *categoryRepo) GetCategory(ctx context.Context, id int64) (*biz.AdminCategory, error) {
	reply, err := r.data.cc.GetCategory(ctx, &v1.GetCategoryRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return protoCategoryToBiz(reply.Item), nil
}

// CreateCategory 创建分类
func (r *categoryRepo) CreateCategory(ctx context.Context, in *biz.AdminCategory) (int64, error) {
	req := bizCategoryToProto(in)
	req.ParentId = in.ParentID
	reply, err := r.data.cc.CreateCategory(ctx, &v1.CreateCategoryRequest{
		Name:       req.Name,
		Slug:       req.Slug,
		Icon:       req.Icon,
		Sort:       req.Sort,
		Status:     req.Status,
		Showcase:   req.Showcase,
		Image:      req.Image,
		SeoTitle:   req.SeoTitle,
		SeoKeywords: req.SeoKeywords,
		SeoDesc:    req.SeoDesc,
		ParentId:   req.ParentId,
	})
	if err != nil {
		return 0, err
	}
	return reply.Id, nil
}

// UpdateCategory 更新分类
func (r *categoryRepo) UpdateCategory(ctx context.Context, in *biz.AdminCategory) error {
	_, err := r.data.cc.UpdateCategory(ctx, &v1.UpdateCategoryRequest{
		Id:         in.ID,
		Name:       in.Name,
		Slug:       in.Slug,
		Icon:       in.Icon,
		Sort:       in.Sort,
		Status:     in.Status,
		Showcase:   in.Showcase,
		Image:      in.Image,
		SeoTitle:   in.SeoTitle,
		SeoKeywords: in.SeoKeywords,
		SeoDesc:    in.SeoDesc,
		ParentId:   in.ParentID,
	})
	return err
}

// DeleteCategory 删除分类
func (r *categoryRepo) DeleteCategory(ctx context.Context, id int64) error {
	_, err := r.data.cc.DeleteCategory(ctx, &v1.DeleteCategoryRequest{Id: id})
	return err
}

// UpdateCategorySort 更新排序
func (r *categoryRepo) UpdateCategorySort(ctx context.Context, id int64, sort int32) error {
	_, err := r.data.cc.UpdateCategorySort(ctx, &v1.UpdateCategorySortRequest{Id: id, Sort: sort})
	return err
}

// UpdateCategoryStatus 更新状态
func (r *categoryRepo) UpdateCategoryStatus(ctx context.Context, id int64, status bool) error {
	_, err := r.data.cc.UpdateCategoryStatus(ctx, &v1.UpdateCategoryStatusRequest{Id: id, Status: status})
	return err
}

// ToggleCategoryShowcase 切换首页展示
func (r *categoryRepo) ToggleCategoryShowcase(ctx context.Context, id int64, showcase bool) error {
	_, err := r.data.cc.ToggleCategoryShowcase(ctx, &v1.ToggleCategoryShowcaseRequest{Id: id, Showcase: showcase})
	return err
}

// BatchCategoryOperation 批量操作分类
func (r *categoryRepo) BatchCategoryOperation(ctx context.Context, action string, ids []int64) (int32, error) {
	count := int32(0)
	for _, id := range ids {
		var err error
		switch action {
		case "删除":
			_, err = r.data.cc.DeleteCategory(ctx, &v1.DeleteCategoryRequest{Id: id})
		case "启用":
			_, err = r.data.cc.UpdateCategoryStatus(ctx, &v1.UpdateCategoryStatusRequest{Id: id, Status: true})
		case "禁用":
			_, err = r.data.cc.UpdateCategoryStatus(ctx, &v1.UpdateCategoryStatusRequest{Id: id, Status: false})
		default:
			continue
		}
		if err != nil {
			r.log.Warnf("batch operation %s failed for id %d: %v", action, id, err)
			continue
		}
		count++
	}
	return count, nil
}
