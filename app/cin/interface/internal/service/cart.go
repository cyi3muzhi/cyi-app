package service

import (
	"context"

	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/pkg/kratos/service"
)

// AddCartItem 加入购物车（同商品同规格数量自动累加）
func (s *CinInterface) AddCartItem(ctx context.Context, req *v1.AddCartItemRequest) (*v1.AddCartItemReply, error) {
	// 获取当前登录用户 ID
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.cartUC.AddItem(ctx, userID, req.ProductId, req.SkuId, req.Quantity); err != nil {
		s.log.Errorf("加入购物车失败, userID=%d, productID=%d, skuID=%d: %v", userID, req.ProductId, req.SkuId, err)
		return nil, err
	}
	return &v1.AddCartItemReply{Success: true}, nil
}

// GetCart 获取购物车（含商品名称/价格/图片/库存等信息）
func (s *CinInterface) GetCart(ctx context.Context, _ *v1.GetCartRequest) (*v1.GetCartReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.cartUC.GetCart(ctx, userID)
	if err != nil {
		s.log.Errorf("获取购物车失败, userID=%d: %v", userID, err)
		return nil, err
	}
	reply := &v1.GetCartReply{Items: make([]*v1.CartProductItem, 0, len(items))}
	for _, it := range items {
		reply.Items = append(reply.Items, &v1.CartProductItem{
			ProductId: it.ProductID,
			SkuId:     it.SkuID,
			Name:      it.Name,
			Price:     it.Price,
			Quantity:  it.Quantity,
			Image:     it.Image,
			Stock:     it.Stock,
			Selected:  it.Selected,
			Sku:       it.Sku,
			Status:    it.Status,
		})
	}
	return reply, nil
}

// UpdateCartItem 修改购物车商品数量
func (s *CinInterface) UpdateCartItem(ctx context.Context, req *v1.UpdateCartItemRequest) (*v1.UpdateCartItemReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.cartUC.UpdateItem(ctx, userID, req.ProductId, req.SkuId, req.Quantity); err != nil {
		s.log.Errorf("修改购物车数量失败, userID=%d, productID=%d, skuID=%d: %v", userID, req.ProductId, req.SkuId, err)
		return nil, err
	}
	return &v1.UpdateCartItemReply{Success: true}, nil
}

// ToggleCartItem 选中/取消选中购物车商品
func (s *CinInterface) ToggleCartItem(ctx context.Context, req *v1.ToggleCartItemRequest) (*v1.ToggleCartItemReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.cartUC.ToggleItem(ctx, userID, req.ProductId, req.SkuId, req.Selected); err != nil {
		s.log.Errorf("切换购物车选中状态失败, userID=%d, productID=%d, skuID=%d: %v", userID, req.ProductId, req.SkuId, err)
		return nil, err
	}
	return &v1.ToggleCartItemReply{Success: true}, nil
}

// SelectAllCart 全选/取消全选购物车
func (s *CinInterface) SelectAllCart(ctx context.Context, req *v1.SelectAllCartRequest) (*v1.SelectAllCartReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.cartUC.SelectAll(ctx, userID, req.Selected); err != nil {
		s.log.Errorf("购物车全选失败, userID=%d: %v", userID, err)
		return nil, err
	}
	return &v1.SelectAllCartReply{Success: true}, nil
}

// DeleteCartItem 删除购物车商品
func (s *CinInterface) DeleteCartItem(ctx context.Context, req *v1.DeleteCartItemRequest) (*v1.DeleteCartItemReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.cartUC.DeleteItem(ctx, userID, req.ProductId, req.SkuId); err != nil {
		s.log.Errorf("删除购物车商品失败, userID=%d, productID=%d, skuID=%d: %v", userID, req.ProductId, req.SkuId, err)
		return nil, err
	}
	return &v1.DeleteCartItemReply{Success: true}, nil
}

// ClearCart 清空购物车
func (s *CinInterface) ClearCart(ctx context.Context, _ *v1.ClearCartRequest) (*v1.ClearCartReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.cartUC.Clear(ctx, userID); err != nil {
		s.log.Errorf("清空购物车失败, userID=%d: %v", userID, err)
		return nil, err
	}
	return &v1.ClearCartReply{Success: true}, nil
}
