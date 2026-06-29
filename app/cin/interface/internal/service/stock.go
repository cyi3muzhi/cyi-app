package service

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	"cyi-app/pkg/util/convert"
)

func (s *CinInterface) UpdateStock(ctx context.Context, req *v1.UpdateStockRequest) (*v1.UpdateStockReply, error) {
	in := &biz.Stock{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	if err := s.sc.UpdateStock(ctx, in); err != nil {
		return nil, err
	}
	return &v1.UpdateStockReply{
		Success: true,
	}, nil
}
