package service

import (
	"context"
	"cyi-app/app/nats/internal/biz"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

// OrderService 订单服务（NATS消息处理器）
type OrderService struct {
	uc  *biz.OrderUseCase
	log *log.Helper
}

// NewOrderService 创建订单服务
func NewOrderService(uc *biz.OrderUseCase, logger log.Logger) *OrderService {
	return &OrderService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// CreateOrder 处理订单创建事件
func (s *OrderService) CreateOrder(ctx context.Context, msg []byte) error {
	fmt.Println("service CreateOrder, msg:", string(msg))
	return s.uc.CreateOrder(ctx)
}

// HandleOrderPaid 处理订单支付事件
func (s *OrderService) HandleOrderPaid(ctx context.Context, msg []byte) error {
	fmt.Println("service HandleOrderPaid, msg:", string(msg))
	return nil
}

// HandleOrderShipped 处理订单发货事件
func (s *OrderService) HandleOrderShipped(ctx context.Context, msg []byte) error {
	fmt.Println("service HandleOrderShipped, msg:", string(msg))
	return nil
}

// HandleOrderCancelled 处理订单取消事件
func (s *OrderService) HandleOrderCancelled(ctx context.Context, msg []byte) error {
	fmt.Println("service HandleOrderCancelled, msg:", string(msg))
	return nil
}

// HandleOrderRefunded 处理订单退款事件
func (s *OrderService) HandleOrderRefunded(ctx context.Context, msg []byte) error {
	fmt.Println("service HandleOrderRefunded, msg:", string(msg))
	return nil
}
