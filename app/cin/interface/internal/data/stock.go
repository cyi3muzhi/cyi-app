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

func (r *stockRepo) ReserveStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	resp, err := r.data.sc.ReserveStock(ctx, &productV1.ReserveStockRequest{
		ProductId: productID,
		Count:     count,
		Version:   version,
	})
	if err != nil {
		return 0, err
	}
	return resp.NewVersion, nil
}

func (r *stockRepo) GetStock(ctx context.Context, productID int64) (*biz.Stock, error) {
	resp, err := r.data.sc.GetStock(ctx, &productV1.GetStockRequest{
		ProductId: productID,
	})
	if err != nil {
		return nil, err
	}
	return &biz.Stock{
		ProductID: resp.ProductId,
		Total:     resp.Total,
		Available: resp.Available,
		Reserved:  resp.Reserved,
		Version:   resp.Version,
	}, nil
}
