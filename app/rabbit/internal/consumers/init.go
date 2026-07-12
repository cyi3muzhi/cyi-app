package consumers

import (
	"cyi-app/app/rabbit/internal/service"
	"cyi-app/pkg/rmq"

	"github.com/google/wire"
)

// ProviderSet is consumers providers.
var ProviderSet = wire.NewSet(
	NewConsumer,
	NewSimpleConsumer,
	NewTopicConsumer,
)

type Consumer struct {
	simpleConsumer *SimpleConsumer
	topicConsumer  *TopicConsumer

	orderSrv *service.OrderService
}


func NewConsumer(simpleConsumer *SimpleConsumer, topicConsumer *TopicConsumer, orderSrv *service.OrderService) *Consumer {
	return &Consumer{simpleConsumer: simpleConsumer, topicConsumer: topicConsumer, orderSrv: orderSrv}
}

func (c *Consumer) Start() error {
	c.SetConsumers()
	go c.simpleConsumer.Start()
	go c.topicConsumer.Start()
	return nil
}

func (c *Consumer) SetConsumers() {
	c.simpleConsumer.SetConsumer(rmq.QueueOrderCreated, c.orderSrv.CreateOrder)

	c.topicConsumer.SetConsumer(rmq.ExchangeLogs, rmq.QueueLogsAll, "order.*", c.orderSrv.CreateOrderLog)
}
