package service

import (
	"context"
	"strconv"

	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	"cyi-app/pkg/util/convert"
)

func (s *CinInterface) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductReply, error) {
	in := &biz.Product{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	if err := s.pc.CreateProduct(ctx, in); err != nil {
		return nil, err
	}
	return &v1.CreateProductReply{
		ProductId: in.ID,
	}, nil
}

func (s *CinInterface) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductReply, error) {
	in := &biz.Product{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	in.ID = req.ProductId
	if err := s.pc.UpdateProduct(ctx, in); err != nil {
		return nil, err
	}
	return &v1.UpdateProductReply{
		Success: true,
	}, nil
}

func (s *CinInterface) ListProduct(ctx context.Context, req *v1.ListProductRequest) (*v1.ListProductReply, error) {
	categoryID, _ := strconv.ParseInt(req.Category, 10, 64)
	bizReq := &biz.ListProductRequest{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Keyword:    req.Keyword,
		CategoryID: categoryID,
		Status:     "1",
	}
	items, total, err := s.pc.ListProduct(ctx, bizReq)
	if err != nil {
		return nil, err
	}
	productItems := make([]*v1.ProductItem, 0, len(items))
	for _, item := range items {
		productItems = append(productItems, convertProductItem(item))
	}
	return &v1.ListProductReply{
		Items: productItems,
		Total: total,
	}, nil
}

func (s *CinInterface) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.GetProductReply, error) {
	product, err := s.pc.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetProductReply{
		Product: convertProductItem(product),
	}, nil
}

func convertProductItem(item *biz.Product) *v1.ProductItem {
	if item == nil {
		return nil
	}
	specs := make([]*v1.SpecGroupInfo, 0, len(item.Specs))
	for _, s := range item.Specs {
		specs = append(specs, &v1.SpecGroupInfo{
			Id:     s.ID,
			Name:   s.Name,
			Values: s.Values,
		})
	}
	skus := make([]*v1.SkuInfo, 0, len(item.Skus))
	for _, s := range item.Skus {
		specMap := make(map[string]string)
		for k, v := range s.Spec {
			specMap[k] = v
		}
		skus = append(skus, &v1.SkuInfo{
			Id:      s.ID,
			SkuCode: s.SkuCode,
			Price:   s.Price,
			Stock:   s.Stock,
			Sales:   s.Sales,
			Spec:    specMap,
			Image:   s.Image,
		})
	}
	return &v1.ProductItem{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		Stock:       item.Stock,
		Category:    item.Category,
		Status:      strconv.Itoa(int(item.Status)),
		Image:       item.Image,
		Sales:       item.Sales,
		CategoryId:  item.CategoryID,
		Images:      item.Images,
		Specs:       specs,
		Skus:        skus,
	}
}
