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
	return &biz.Product{
		ID:          resp.Product.ProductId,
		Name:        resp.Product.Name,
		Description: resp.Product.Description,
		Price:       resp.Product.Price,
		Status:      resp.Product.Status,
	}, nil
}
