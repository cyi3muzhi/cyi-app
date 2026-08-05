package data

import (
	"context"
	"cyi-app/app/cin/interface/internal/biz"

	productV1 "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type productRepo struct {
	log  *log.Helper
	data *Data
}

func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

func (r *productRepo) CreateProduct(ctx context.Context, in *biz.Product) error {
	resp, err := r.data.pc.CreateProduct(ctx, &productV1.CreateProductRequest{
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
	})
	if err != nil {
		return err
	}
	in.ID = resp.ProductId
	return nil
}

func (r *productRepo) UpdateProduct(ctx context.Context, in *biz.Product) error {
	_, err := r.data.pc.UpdateProduct(ctx, &productV1.UpdateProductRequest{
		ProductId:   in.ID,
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
	})
	return err
}

func (r *productRepo) GetProduct(ctx context.Context, productID int64) (*biz.Product, error) {
	resp, err := r.data.pc.GetProduct(ctx, &productV1.GetProductRequest{
		ProductId: productID,
	})
	if err != nil {
		return nil, err
	}
	return convertProductItem(resp.Product), nil
}

func (r *productRepo) ListProduct(ctx context.Context, req *biz.ListProductRequest) ([]*biz.Product, int32, error) {
	resp, err := r.data.pc.ListProduct(ctx, &productV1.ListProductRequest{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Keyword:    req.Keyword,
		CategoryId: req.CategoryID,
		Status:     req.Status,
	})
	if err != nil {
		return nil, 0, err
	}
	items := make([]*biz.Product, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, convertProductItem(item))
	}
	return items, resp.Total, nil
}

// SuggestProduct 搜索联想：代理调用商品服务的 SuggestProduct rpc
// 返回匹配关键词的商品名列表（去重）
func (r *productRepo) SuggestProduct(ctx context.Context, keyword string, limit int32) ([]string, error) {
	resp, err := r.data.pc.SuggestProduct(ctx, &productV1.SuggestProductRequest{
		Keyword: keyword,
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}
	items := resp.Items
	if items == nil {
		items = []string{}
	}
	return items, nil
}

func convertProductItem(item *productV1.ProductItem) *biz.Product {
	if item == nil {
		return nil
	}
	images := item.Images
	if len(images) == 0 {
		images = []string{item.Image}
	}
	return &biz.Product{
		ID:          item.ProductId,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		Stock:       getDefaultStock(item.Skus),
		Category:    item.Category,
		CategoryID:  item.CategoryId,
		Status:      item.Status,
		Image:       images[0],
		Sales:       item.Sales,
		Images:      images,
		Specs:       convertSpecGroups(item.Specs),
		Skus:        convertSkus(item.Skus),
	}
}

func convertSpecGroups(specs []*productV1.SpecGroupInfo) []*biz.SpecGroupInfo {
	result := make([]*biz.SpecGroupInfo, 0, len(specs))
	for _, s := range specs {
		result = append(result, &biz.SpecGroupInfo{
			ID:     s.Id,
			Name:   s.Name,
			Values: s.Values,
		})
	}
	return result
}

func convertSkus(skus []*productV1.SkuInfo) []*biz.SkuInfo {
	result := make([]*biz.SkuInfo, 0, len(skus))
	for _, s := range skus {
		specMap := make(map[string]string)
		for k, v := range s.Spec {
			specMap[k] = v
		}
		result = append(result, &biz.SkuInfo{
			ID:      s.Id,
			SkuCode: s.SkuCode,
			Price:   s.Price,
			Stock:   s.Stock,
			Sales:   s.Sales,
			Spec:    specMap,
			Image:   s.Image,
		})
	}
	return result
}

func getDefaultStock(skus []*productV1.SkuInfo) int32 {
	if len(skus) == 0 {
		return 0
	}
	var total int32
	for _, sku := range skus {
		total += sku.Stock
	}
	return total
}
