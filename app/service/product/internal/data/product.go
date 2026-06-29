package data

import (
	"context"
	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type productRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

func (r *productRepo) CreateProduct(ctx context.Context, in *biz.Product) error {
	product := &mod.Product{}
	convert.Convert(in, product)
	if err := curd.Product.WithContext(ctx).Create(product); err != nil {
		return err
	}
	in.ID = product.ID
	return nil
}

func (r *productRepo) UpdateProduct(ctx context.Context, in *biz.Product) error {
	product := &mod.Product{}
	convert.Convert(in, product)
	if _, err := curd.Product.Debug().WithContext(ctx).Where(curd.Product.ID.Eq(in.ID)).Updates(product); err != nil {
		return err
	}
	return nil
}

func (r *productRepo) GetProduct(ctx context.Context, productID int64) (*biz.Product, error) {
	product, err := curd.Product.WithContext(ctx).Where(curd.Product.ID.Eq(productID)).First()
	if err != nil {
		return nil, err
	}
	out := &biz.Product{}
	convert.Convert(product, out)
	return out, nil
}
