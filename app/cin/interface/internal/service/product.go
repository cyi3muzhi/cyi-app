package service

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	"cyi-app/pkg/util/convert"
)

func (s *CinInterface) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductReply, error) {
	in := &biz.Product{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	if err := s.pc.CreateProduct(ctx, in); err != nil {
		return nil, err
	}
	return &v1.CreateProductReply{
		ProductId: in.ID,
	}, nil
}

func (s *CinInterface) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductReply, error) {
	in := &biz.Product{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	if err := s.pc.UpdateProduct(ctx, in); err != nil {
		return nil, err
	}
	return &v1.UpdateProductReply{
		Success: true,
	}, nil
}
