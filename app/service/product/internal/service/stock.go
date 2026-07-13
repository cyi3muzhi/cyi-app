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

func (s *StockService) ReserveStock(ctx context.Context, req *v1.ReserveStockRequest) (*v1.ReserveStockReply, error) {
	newVersion, err := s.uc.ReserveStock(ctx, req.ProductId, req.Count, req.Version)
	if err != nil {
		return nil, err
	}
	return &v1.ReserveStockReply{
		Success:    true,
		NewVersion: newVersion,
	}, nil
}

func (s *StockService) DeductStock(ctx context.Context, req *v1.DeductStockRequest) (*v1.DeductStockReply, error) {
	newVersion, err := s.uc.DeductStock(ctx, req.ProductId, req.Count, req.Version)
	if err != nil {
		return nil, err
	}
	return &v1.DeductStockReply{
		Success:    true,
		NewVersion: newVersion,
	}, nil
}

func (s *StockService) ReleaseStock(ctx context.Context, req *v1.ReleaseStockRequest) (*v1.ReleaseStockReply, error) {
	newVersion, err := s.uc.ReleaseStock(ctx, req.ProductId, req.Count, req.Version)
	if err != nil {
		return nil, err
	}
	return &v1.ReleaseStockReply{
		Success:    true,
		NewVersion: newVersion,
	}, nil
}

func (s *StockService) AddStock(ctx context.Context, req *v1.AddStockRequest) (*v1.AddStockReply, error) {
	newVersion, err := s.uc.AddStock(ctx, req.ProductId, req.Count, req.Version)
	if err != nil {
		return nil, err
	}
	return &v1.AddStockReply{
		Success:    true,
		NewVersion: newVersion,
	}, nil
}

func (s *StockService) DeleteStock(ctx context.Context, req *v1.DeleteStockRequest) (*v1.DeleteStockReply, error) {
	return &v1.DeleteStockReply{}, nil
}

func (s *StockService) GetStock(ctx context.Context, req *v1.GetStockRequest) (*v1.GetStockReply, error) {
	stock, err := s.uc.GetStock(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}
	return &v1.GetStockReply{
		ProductId: stock.ProductID,
		Total:     stock.Total,
		Available: stock.Available,
		Reserved:  stock.Reserved,
		Version:   stock.Version,
	}, nil
}

func (s *StockService) ListStock(ctx context.Context, req *v1.ListStockRequest) (*v1.ListStockReply, error) {
	return &v1.ListStockReply{}, nil
}
