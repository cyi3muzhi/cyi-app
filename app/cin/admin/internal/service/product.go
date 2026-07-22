package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

// ListProduct 获取商品列表
func (s *AdminService) ListProduct(ctx context.Context, req *adminv1.ListProductRequest) (*adminv1.ListProductReply, error) {
	// 解包 Int32Value：nil 表示未设置（不筛选），否则取实际值
	var stockEQ int32 = -1
	if sw := req.GetStockEq(); sw != nil {
		stockEQ = sw.GetValue()
	}
	items, total, err := s.pc.ListProduct(ctx, req.Page, req.PageSize, req.Keyword, req.CategoryId, req.Status, stockEQ)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*adminv1.ProductItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, bizToAdminProtoProduct(item))
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
		CategoryID:  req.CategoryId,
		Price:       req.Price,
		Stock:       req.Stock,
		Image:       req.Image,
		Images:      req.GetImages(),
		SKU:         req.Sku,
		Description: req.Description,
		Specs:       adminProtoToBizSpecGroups(req.Specs),
		Skus:        adminProtoToBizSkus(req.Skus),
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
		CategoryID:  req.CategoryId,
		Price:       req.Price,
		Stock:       req.Stock,
		Image:       req.Image,
		Images:      req.GetImages(),
		SKU:         req.Sku,
		Status:      req.Status,
		Description: req.Description,
		Specs:       adminProtoToBizSpecGroups(req.Specs),
		Skus:        adminProtoToBizSkus(req.Skus),
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
		Product: bizToAdminProtoProduct(product),
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

// bizToAdminProtoProduct 将 biz Product 转换为 admin proto ProductItem
func bizToAdminProtoProduct(product *biz.Product) *adminv1.ProductItem {
	return &adminv1.ProductItem{
		Id:          product.ID,
		Name:        product.Name,
		CategoryId:  product.CategoryID,
		Price:       product.Price,
		Stock:       product.Stock,
		Sales:       product.Sales,
		Status:      product.Status,
		Image:       product.Image,
		Images:      product.Images,
		Sku:         product.SKU,
		Specs:       bizToAdminProtoSpecGroups(product.Specs),
		Skus:        bizToAdminProtoSkus(product.Skus),
		Description: product.Description,
	}
}

// bizToAdminProtoSpecGroups 将 biz SpecGroup 列表转换为 admin proto SpecGroupItem 列表
func bizToAdminProtoSpecGroups(specs []*biz.SpecGroup) []*adminv1.SpecGroupItem {
	if len(specs) == 0 {
		return nil
	}
	result := make([]*adminv1.SpecGroupItem, 0, len(specs))
	for _, s := range specs {
		result = append(result, &adminv1.SpecGroupItem{
			Id:     s.ID,
			Name:   s.Name,
			Values: s.Values,
		})
	}
	return result
}

// bizToAdminProtoSkus 将 biz SkuItem 列表转换为 admin proto SkuItem 列表
func bizToAdminProtoSkus(skus []*biz.SkuItem) []*adminv1.SkuItem {
	if len(skus) == 0 {
		return nil
	}
	result := make([]*adminv1.SkuItem, 0, len(skus))
	for _, s := range skus {
		result = append(result, &adminv1.SkuItem{
			Id:      s.ID,
			SkuCode: s.SkuCode,
			Price:   s.Price,
			Stock:   s.Stock,
			Image:   s.Image,
			Spec:    s.Spec,
		})
	}
	return result
}

// adminProtoToBizSpecGroups 将 admin proto SpecGroupItem 列表转换为 biz SpecGroup 列表
func adminProtoToBizSpecGroups(specs []*adminv1.SpecGroupItem) []*biz.SpecGroup {
	if len(specs) == 0 {
		return nil
	}
	result := make([]*biz.SpecGroup, 0, len(specs))
	for _, s := range specs {
		result = append(result, &biz.SpecGroup{
			ID:     s.Id,
			Name:   s.Name,
			Values: s.Values,
		})
	}
	return result
}

// adminProtoToBizSkus 将 admin proto SkuItem 列表转换为 biz SkuItem 列表
func adminProtoToBizSkus(skus []*adminv1.SkuItem) []*biz.SkuItem {
	if len(skus) == 0 {
		return nil
	}
	result := make([]*biz.SkuItem, 0, len(skus))
	for _, s := range skus {
		result = append(result, &biz.SkuItem{
			ID:      s.Id,
			SkuCode: s.SkuCode,
			Price:   s.Price,
			Stock:   s.Stock,
			Image:   s.GetImage(),
			Spec:    s.Spec,
		})
	}
	return result
}
