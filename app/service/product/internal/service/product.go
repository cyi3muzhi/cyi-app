package service

import (
	"context"
	"errors"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/pkg/util/convert"

	v1 "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// ProductService 商品服务
type ProductService struct {
	v1.UnimplementedProductServer
	uc  *biz.ProductUseCase
	log log.Logger
}

// NewProductService 创建商品服务
func NewProductService(uc *biz.ProductUseCase, logger log.Logger) *ProductService {
	return &ProductService{
		uc:  uc,
		log: logger,
	}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductReply, error) {
	in := &biz.Product{}
	convert.Convert(req, in)
	in.Description = req.Description // Convert 忽略 error 时可能未映射，显式赋值
	in.CategoryID = req.GetCategoryId()
	in.Image = req.Image
	in.Images = req.GetImages()
	// 转换规格组
	in.SpecGroups = specGroupInfoListToBiz(req.Specs)
	// 转换SKU
	in.Skus = skuInfoListToBiz(req.Skus)
	if err := s.uc.CreateProduct(ctx, in); err != nil {
		return nil, convertBizError(err)
	}
	return &v1.CreateProductReply{
		ProductId: in.ID,
	}, nil
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductReply, error) {
	in := &biz.Product{
		ID:          req.ProductId,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.GetCategoryId(),
		Image:       req.Image,
		Images:      req.GetImages(),
		Status:      biz.ProductStatus(req.Status),
		SpecGroups:  specGroupInfoListToBiz(req.Specs),
		Skus:        skuInfoListToBiz(req.Skus),
	}
	if err := s.uc.UpdateProduct(ctx, in); err != nil {
		return nil, err
	}
	return &v1.UpdateProductReply{
		Success: true,
	}, nil
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*v1.DeleteProductReply, error) {
	if err := s.uc.DeleteProduct(ctx, req.ProductId); err != nil {
		return nil, err
	}
	return &v1.DeleteProductReply{
		Success: true,
	}, nil
}

// GetProduct 获取商品详情
func (s *ProductService) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.GetProductReply, error) {
	product, err := s.uc.GetProduct(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}
	return &v1.GetProductReply{
		Product: bizProductToProductItem(product),
	}, nil
}

// ListProduct 获取商品列表
func (s *ProductService) ListProduct(ctx context.Context, req *v1.ListProductRequest) (*v1.ListProductReply, error) {
	// 解包 Int32Value：nil 表示未设置（不筛选），否则取实际值
	var stockEQ int32 = -1
	if sw := req.GetStockEq(); sw != nil {
		stockEQ = sw.GetValue()
	}
	items, total, err := s.uc.ListProduct(ctx, req.Page, req.PageSize, req.Keyword, req.GetCategoryId(), req.Status, stockEQ)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*v1.ProductItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, bizProductToProductItem(item))
	}
	return &v1.ListProductReply{
		Items: replyItems,
		Total: total,
	}, nil
}

// ====== 转换辅助函数 ======

// specGroupInfoListToBiz 将 proto 规格组列表转换为 biz 实体列表
func specGroupInfoListToBiz(specs []*v1.SpecGroupInfo) []*biz.SpecGroup {
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

// specGroupToProto 将 biz 规格组实体转换为 proto 消息
func specGroupToProto(sg *biz.SpecGroup) *v1.SpecGroupInfo {
	return &v1.SpecGroupInfo{
		Id:     sg.ID,
		Name:   sg.Name,
		Values: sg.Values,
	}
}

// specGroupListToProto 将 biz 规格组列表转换为 proto 列表
func specGroupListToProto(sgs []*biz.SpecGroup) []*v1.SpecGroupInfo {
	if len(sgs) == 0 {
		return nil
	}
	result := make([]*v1.SpecGroupInfo, 0, len(sgs))
	for _, sg := range sgs {
		result = append(result, specGroupToProto(sg))
	}
	return result
}

// skuInfoListToBiz 将 proto SKU 列表转换为 biz 实体列表
func skuInfoListToBiz(skus []*v1.SkuInfo) []*biz.Sku {
	if len(skus) == 0 {
		return nil
	}
	result := make([]*biz.Sku, 0, len(skus))
	for _, s := range skus {
		result = append(result, &biz.Sku{
			ID:      s.Id,
			SkuCode: s.SkuCode,
			Price:   s.Price,
			Stock:   s.Stock,
			Sales:   s.Sales,
			Spec:    s.Spec,
			Image:   s.GetImage(),
		})
	}
	return result
}

// skuToProto 将 biz SKU 实体转换为 proto 消息
func skuToProto(s *biz.Sku) *v1.SkuInfo {
	return &v1.SkuInfo{
		Id:      s.ID,
		SkuCode: s.SkuCode,
		Price:   s.Price,
		Stock:   s.Stock,
		Sales:   s.Sales,
		Spec:    s.Spec,
		Image:   s.Image,
	}
}

// skuListToProto 将 biz SKU 列表转换为 proto 列表
func skuListToProto(skus []*biz.Sku) []*v1.SkuInfo {
	if len(skus) == 0 {
		return nil
	}
	result := make([]*v1.SkuInfo, 0, len(skus))
	for _, s := range skus {
		result = append(result, skuToProto(s))
	}
	return result
}

// convertBizError 将 biz 层错误转换为 proto 错误
func convertBizError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, biz.ErrCreateProductFailed) {
		return v1.ErrorCreateProductFailed("%s", err.Error())
	}
	if errors.Is(err, biz.ErrCreateStockFailed) {
		return v1.ErrorCreateStockFailed("%s", err.Error())
	}
	return err
}

// bizProductToProductItem 将 biz.Product 转换为 proto ProductItem
func bizProductToProductItem(p *biz.Product) *v1.ProductItem {
	return &v1.ProductItem{
		ProductId:   p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Status:      p.Status.Value(),
		Sales:       p.Sales,
		CategoryId:   p.CategoryID,
		Image:       p.Image,
		Images:      p.Images,
		Specs:       specGroupListToProto(p.SpecGroups),
		Skus:        skuListToProto(p.Skus),
	}
}
