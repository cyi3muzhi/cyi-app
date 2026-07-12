package consumers

import (
	"cyi-app/app/nats/internal/service"

	"github.com/google/wire"
)

// ProviderSet is consumers providers.
var ProviderSet = wire.NewSet(
	NewConsumer,
	NewNatsConsumer,
)

// Consumer NATS消费者编排器
type Consumer struct {
	natsConsumer *NatsConsumer
	orderSrv     *service.OrderService
}

// NewConsumer 创建消费者编排器
func NewConsumer(natsConsumer *NatsConsumer, orderSrv *service.OrderService) *Consumer {
	return &Consumer{natsConsumer: natsConsumer, orderSrv: orderSrv}
}

// Start 启动所有NATS消费者
func (c *Consumer) Start() error {
	c.SetConsumers()
	return c.natsConsumer.Start()
}

// SetConsumers 注册消费者处理器
func (c *Consumer) SetConsumers() {
	c.natsConsumer.Subscribe("order.created", c.orderSrv.CreateOrder)
	c.natsConsumer.Subscribe("order.paid", c.orderSrv.HandleOrderPaid)
	c.natsConsumer.Subscribe("order.shipped", c.orderSrv.HandleOrderShipped)
	c.natsConsumer.Subscribe("order.cancelled", c.orderSrv.HandleOrderCancelled)
	c.natsConsumer.Subscribe("order.refunded", c.orderSrv.HandleOrderRefunded)
}
