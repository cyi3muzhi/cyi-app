package data

import (
	"context"

	"cyi-app/app/service/order/internal/biz"
	productRpc "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type stockRepo struct {
	data *Data
	log  *log.Helper
}

func NewStockRepo(data *Data, logger log.Logger) biz.StockRepo {
	return &stockRepo{data: data, log: log.NewHelper(logger)}
}

func (r *stockRepo) GetStock(ctx context.Context, productID int64) (*productRpc.GetStockReply, error) {
	reply, err := r.data.sc.GetStock(ctx, &productRpc.GetStockRequest{ProductId: productID})
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (r *stockRepo) ReserveStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	reply, err := r.data.sc.ReserveStock(ctx, &productRpc.ReserveStockRequest{
		ProductId: productID,
		Count:     count,
		Version:   version,
	})
	if err != nil {
		return 0, err
	}
	return reply.NewVersion, nil
}

func (r *stockRepo) DeductStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	reply, err := r.data.sc.DeductStock(ctx, &productRpc.DeductStockRequest{
		ProductId: productID,
		Count:     count,
		Version:   version,
	})
	if err != nil {
		return 0, err
	}
	return reply.NewVersion, nil
}

func (r *stockRepo) ReleaseStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	reply, err := r.data.sc.ReleaseStock(ctx, &productRpc.ReleaseStockRequest{
		ProductId: productID,
		Count:     count,
		Version:   version,
	})
	if err != nil {
		return 0, err
	}
	return reply.NewVersion, nil
}

func (r *stockRepo) AddStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	reply, err := r.data.sc.AddStock(ctx, &productRpc.AddStockRequest{
		ProductId: productID,
		Count:     count,
		Version:   version,
	})
	if err != nil {
		return 0, err
	}
	return reply.NewVersion, nil
}
