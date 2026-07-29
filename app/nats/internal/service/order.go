package service

import (
	"context"
	"cyi-app/app/nats/internal/biz"

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
	s.log.WithContext(ctx).Infof("收到订单创建事件: %s", string(msg))
	return s.uc.CreateOrder(ctx)
}

// HandleOrderPaid 处理订单支付事件
// TODO: 待完善订单支付后的业务逻辑
func (s *OrderService) HandleOrderPaid(ctx context.Context, msg []byte) error {
	s.log.WithContext(ctx).Infof("收到订单支付事件: %s", string(msg))
	return nil
}

// HandleOrderShipped 处理订单发货事件
// TODO: 待完善订单发货后的业务逻辑
func (s *OrderService) HandleOrderShipped(ctx context.Context, msg []byte) error {
	s.log.WithContext(ctx).Infof("收到订单发货事件: %s", string(msg))
	return nil
}

// HandleOrderCancelled 处理订单取消事件
// TODO: 待完善订单取消后的业务逻辑
func (s *OrderService) HandleOrderCancelled(ctx context.Context, msg []byte) error {
	s.log.WithContext(ctx).Infof("收到订单取消事件: %s", string(msg))
	return nil
}

// HandleOrderRefunded 处理订单退款事件
// TODO: 待完善订单退款后的业务逻辑
func (s *OrderService) HandleOrderRefunded(ctx context.Context, msg []byte) error {
	s.log.WithContext(ctx).Infof("收到订单退款事件: %s", string(msg))
	return nil
}
