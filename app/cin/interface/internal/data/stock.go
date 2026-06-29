package data

import (
	"context"
	"cyi-app/app/cin/interface/internal/biz"

	productV1 "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type stockRepo struct {
	log  *log.Helper
	data *Data
}

func NewStockRepo(data *Data, logger log.Logger) biz.StockRepo {
	return &stockRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

func (r *stockRepo) UpdateStock(ctx context.Context, in *biz.Stock) error {
	_, err := r.data.sc.UpdateStock(ctx, &productV1.UpdateStockRequest{
		ProductId: in.ProductID,
		Total:     in.Total,
	})
	return err
}
