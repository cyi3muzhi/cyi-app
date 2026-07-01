package data

import (
	"context"

	productRpc "cyi-app/api/service/product/v1"
	"cyi-app/app/service/order/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type productRepo struct {
	data *Data
	log  *log.Helper
}

func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{data: data, log: log.NewHelper(logger)}
}

func (r *productRepo) GetProduct(ctx context.Context, productID int64) (*productRpc.GetProductReply, error) {
	reply, err := r.data.pc.GetProduct(ctx, &productRpc.GetProductRequest{ProductId: productID})
	if err != nil {
		return nil, err
	}
	return reply, nil
}
