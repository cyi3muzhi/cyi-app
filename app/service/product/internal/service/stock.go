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

func (s *StockService) DeleteStock(ctx context.Context, req *v1.DeleteStockRequest) (*v1.DeleteStockReply, error) {
	return &v1.DeleteStockReply{}, nil
}

func (s *StockService) GetStock(ctx context.Context, req *v1.GetStockRequest) (*v1.GetStockReply, error) {
	return &v1.GetStockReply{}, nil
}

func (s *StockService) ListStock(ctx context.Context, req *v1.ListStockRequest) (*v1.ListStockReply, error) {
	return &v1.ListStockReply{}, nil
}
