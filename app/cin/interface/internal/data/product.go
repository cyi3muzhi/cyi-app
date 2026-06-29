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
