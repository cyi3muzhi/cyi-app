package consumers

import (
	"context"
	"cyi-app/pkg/rmq"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/streadway/amqp"
)

type TopicConsumer struct {
	factory  *rmq.QueueFactory
	handlers map[string]map[string]map[string]func(context.Context, []byte) error
	log      *log.Helper
}

func NewTopicConsumer(factory *rmq.QueueFactory, logger log.Logger) *TopicConsumer {
	return &TopicConsumer{factory: factory, log: log.NewHelper(logger)}
}

func (c *TopicConsumer) Start() error {
	for exchange, queueMap := range c.GetConsumers() {
		for queue, bindingKeyMap := range queueMap {
			for bindingKey, handler := range bindingKeyMap {
				fmt.Printf("exchange: %s, queue: %s, bindingKey: %s\n", exchange, queue, bindingKey)
				queue, err := c.factory.CreateTopicQueue(exchange, queue, bindingKey)
				if err != nil {
					return err
				}
				err = queue.Consume(func(d amqp.Delivery) {
					handler(context.Background(), d.Body)
				})
				if err != nil {
					c.log.Errorf("Failed to consume: %v", err)
					return err
				}
			}
		}
	}
	return nil
}

// 获取消费者队列
func (c *TopicConsumer) GetConsumers() map[string]map[string]map[string]func(context.Context, []byte) error {
	return c.handlers
}

func (c *TopicConsumer) SetConsumer(exchange, queue, bindingKey string, handler func(context.Context, []byte) error) error {
	if c.handlers == nil {
		c.handlers = make(map[string]map[string]map[string]func(context.Context, []byte) error)
	}

	// 初始化 exchange 层
	if _, ok := c.handlers[exchange]; !ok {
		c.handlers[exchange] = make(map[string]map[string]func(context.Context, []byte) error)
	}

	// 初始化 queue 层
	if _, ok := c.handlers[exchange][queue]; !ok {
		c.handlers[exchange][queue] = make(map[string]func(context.Context, []byte) error)
	}

	c.handlers[exchange][queue][bindingKey] = handler
	return nil
}
