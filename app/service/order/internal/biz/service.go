package biz

import (
	"context"
	productRpc "cyi-app/api/service/product/v1"
)

// ProductRepo 商品仓储接口（远程调用商品服务）
type ProductRepo interface {
	GetProduct(ctx context.Context, productID int64) (*productRpc.GetProductReply, error)
}

// StockRepo 库存仓储接口（远程调用库存服务）
type StockRepo interface {
	GetStock(ctx context.Context, productID int64) (*productRpc.GetStockReply, error)
	ReserveStock(ctx context.Context, productID int64, count int32, version int32) (int32, error)
	DeductStock(ctx context.Context, productID int64, count int32, version int32) (int32, error)
}
