package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

// ListProduct 获取商品列表
func (s *AdminService) ListProduct(ctx context.Context, req *adminv1.ListProductRequest) (*adminv1.ListProductReply, error) {
	items, total, err := s.pc.ListProduct(ctx, req.Page, req.PageSize, req.Keyword, req.Category, req.Status)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*adminv1.ProductItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, &adminv1.ProductItem{
			Id:       item.ID,
			Name:     item.Name,
			Category: item.Category,
			Price:    item.Price,
			Stock:    item.Stock,
			Sales:    item.Sales,
			Status:   item.Status,
			Image:    item.Image,
			Sku:      item.SKU,
		})
	}
	return &adminv1.ListProductReply{
		Items: replyItems,
		Total: total,
	}, nil
}

// CreateProduct 创建商品
func (s *AdminService) CreateProduct(ctx context.Context, req *adminv1.CreateProductRequest) (*adminv1.CreateProductReply, error) {
	id, err := s.pc.CreateProduct(ctx, &biz.Product{
		Name:        req.Name,
		Category:    req.Category,
		Price:       req.Price,
		Stock:       req.Stock,
		Image:       req.Image,
		SKU:         req.Sku,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return &adminv1.CreateProductReply{
		Id: id,
	}, nil
}

// UpdateProduct 更新商品
func (s *AdminService) UpdateProduct(ctx context.Context, req *adminv1.UpdateProductRequest) (*adminv1.UpdateProductReply, error) {
	err := s.pc.UpdateProduct(ctx, &biz.Product{
		ID:          req.Id,
		Name:        req.Name,
		Category:    req.Category,
		Price:       req.Price,
		Stock:       req.Stock,
		Image:       req.Image,
		SKU:         req.Sku,
		Status:      req.Status,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return &adminv1.UpdateProductReply{
		Success: true,
	}, nil
}

// DeleteProduct 删除商品
func (s *AdminService) DeleteProduct(ctx context.Context, req *adminv1.DeleteProductRequest) (*adminv1.DeleteProductReply, error) {
	err := s.pc.DeleteProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &adminv1.DeleteProductReply{
		Success: true,
	}, nil
}

// GetProduct 获取商品详情
func (s *AdminService) GetProduct(ctx context.Context, req *adminv1.GetProductRequest) (*adminv1.GetProductReply, error) {
	product, err := s.pc.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &adminv1.GetProductReply{
		Product: &adminv1.ProductItem{
			Id:       product.ID,
			Name:     product.Name,
			Category: product.Category,
			Price:    product.Price,
			Stock:    product.Stock,
			Sales:    product.Sales,
			Status:   product.Status,
			Image:    product.Image,
			Sku:      product.SKU,
		},
	}, nil
}

// BatchOperation 批量操作商品
func (s *AdminService) BatchOperation(ctx context.Context, req *adminv1.BatchOperationRequest) (*adminv1.BatchOperationReply, error) {
	affected, err := s.pc.BatchOperation(ctx, req.Action, req.Ids)
	if err != nil {
		return nil, err
	}
	return &adminv1.BatchOperationReply{
		Success:       true,
		AffectedCount: affected,
	}, nil
}
