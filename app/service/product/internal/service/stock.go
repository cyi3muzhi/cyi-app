package service

import (
	"context"

	v1 "cyi-app/api/service/product/v1"
	"cyi-app/app/service/product/internal/biz"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
)

type StockService struct {
	v1.UnimplementedStockServer
	uc  *biz.StockUseCase
	log log.Logger
}

func NewStockService(uc *biz.StockUseCase, logger log.Logger) *StockService {
	return &StockService{
		uc:  uc,
		log: logger,
	}
}

func (s *StockService) CreateStock(ctx context.Context, req *v1.CreateStockRequest) (*v1.CreateStockReply, error) {
	return &v1.CreateStockReply{}, nil
}

func (s *StockService) UpdateStock(ctx context.Context, req *v1.UpdateStockRequest) (*v1.UpdateStockReply, error) {
	in := &biz.Stock{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	if err := s.uc.UpdateStock(ctx, in); err != nil {
		return nil, err
	}
	return &v1.UpdateStockReply{
		Success: true,
	}, nil
}

// UpdateSkuStocks 按 SKU 批量设置库存（管理端按 SKU 明细补货）
func (s *StockService) UpdateSkuStocks(ctx context.Context, req *v1.UpdateSkuStocksRequest) (*v1.UpdateSkuStocksReply, error) {
	items := make(map[int64]int32, len(req.GetItems()))
	for _, it := range req.GetItems() {
		items[it.GetSkuId()] = it.GetTotal()
	}
	if err := s.uc.UpdateSkuStocks(ctx, req.GetProductId(), items); err != nil {
		return nil, err
	}
	return &v1.UpdateSkuStocksReply{
		Success: true,
	}, nil
}

// ReserveStock gRPC 直调入口（管理端/内部调用）：无 DTM 事务上下文，传空幂等键按普通条件更新执行
func (s *StockService) ReserveStock(ctx context.Context, req *v1.ReserveStockRequest) (*v1.ReserveStockReply, error) {
	if err := s.uc.ReserveStock(ctx, biz.StockOpKey{}, req.GetSkuId(), req.Count); err != nil {
		return nil, err
	}
	return &v1.ReserveStockReply{
		Success: true,
	}, nil
}

// DeductStock gRPC 直调入口：无 DTM 事务上下文，传空幂等键按普通条件更新执行
func (s *StockService) DeductStock(ctx context.Context, req *v1.DeductStockRequest) (*v1.DeductStockReply, error) {
	if err := s.uc.DeductStock(ctx, biz.StockOpKey{}, req.GetSkuId(), req.Count); err != nil {
		return nil, err
	}
	return &v1.DeductStockReply{
		Success: true,
	}, nil
}

// ReleaseStock gRPC 直调入口：无 DTM 事务上下文，传空幂等键按普通条件更新执行
func (s *StockService) ReleaseStock(ctx context.Context, req *v1.ReleaseStockRequest) (*v1.ReleaseStockReply, error) {
	if err := s.uc.ReleaseStock(ctx, biz.StockOpKey{}, req.GetSkuId(), req.Count); err != nil {
		return nil, err
	}
	return &v1.ReleaseStockReply{
		Success: true,
	}, nil
}

// AddStock gRPC 直调入口：无 DTM 事务上下文，传空幂等键按普通条件更新执行
func (s *StockService) AddStock(ctx context.Context, req *v1.AddStockRequest) (*v1.AddStockReply, error) {
	if err := s.uc.AddStock(ctx, biz.StockOpKey{}, req.GetSkuId(), req.Count); err != nil {
		return nil, err
	}
	return &v1.AddStockReply{
		Success: true,
	}, nil
}

func (s *StockService) DeleteStock(ctx context.Context, req *v1.DeleteStockRequest) (*v1.DeleteStockReply, error) {
	return &v1.DeleteStockReply{}, nil
}

func (s *StockService) GetStock(ctx context.Context, req *v1.GetStockRequest) (*v1.GetStockReply, error) {
	stock, err := s.uc.GetStock(ctx, req.GetSkuId())
	if err != nil {
		return nil, err
	}
	return &v1.GetStockReply{
		ProductId: stock.ProductID,
		Total:     stock.Total,
		Available: stock.Available,
		Reserved:  stock.Reserved,
	}, nil
}

func (s *StockService) ListStock(ctx context.Context, req *v1.ListStockRequest) (*v1.ListStockReply, error) {
	return &v1.ListStockReply{}, nil
}
