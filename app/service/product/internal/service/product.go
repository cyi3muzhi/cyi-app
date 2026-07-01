package service

import (
	"context"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/pkg/util/convert"

	v1 "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type ProductService struct {
	v1.UnimplementedProductServer
	uc  *biz.ProductUseCase
	log log.Logger
}

func NewProductService(uc *biz.ProductUseCase, logger log.Logger) *ProductService {
	return &ProductService{
		uc:  uc,
		log: logger,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductReply, error) {
	in := &biz.Product{}
	convert.Convert(req, in)
	err := s.uc.CreateProduct(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.CreateProductReply{
		ProductId: in.ID,
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductReply, error) {
	in := &biz.Product{
		ID:          req.ProductId,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}
	if err := s.uc.UpdateProduct(ctx, in); err != nil {
		return nil, err
	}
	return &v1.UpdateProductReply{}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*v1.DeleteProductReply, error) {
	return &v1.DeleteProductReply{}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.GetProductReply, error) {
	product, err := s.uc.GetProduct(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}
	return &v1.GetProductReply{
		ProductId:   product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Status:      product.Status.Value(),
	}, nil
}

func (s *ProductService) ListProduct(ctx context.Context, req *v1.ListProductRequest) (*v1.ListProductReply, error) {
	return &v1.ListProductReply{}, nil
}
