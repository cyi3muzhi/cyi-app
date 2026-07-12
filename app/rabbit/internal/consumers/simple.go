package consumers

import (
	"context"
	"cyi-app/pkg/rmq"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/streadway/amqp"
)

type SimpleConsumer struct {
	factory  *rmq.QueueFactory
	handlers map[string]func(context.Context, []byte) error
	log      *log.Helper
}

func NewSimpleConsumer(factory *rmq.QueueFactory, logger log.Logger) *SimpleConsumer {
	return &SimpleConsumer{factory: factory, log: log.NewHelper(logger)}
}

func (c *SimpleConsumer) Start() error {
	for queue, handler := range c.GetConsumers() {
		queue, err := c.factory.CreateSimpleQueue(queue)
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
	return nil
}

// 获取消费者队列
func (c *SimpleConsumer) GetConsumers() map[string]func(context.Context, []byte) error {
	return c.handlers
}

func (c *SimpleConsumer) SetConsumer(queue string, handler func(context.Context, []byte) error) error {
	if c.handlers == nil {
		c.handlers = make(map[string]func(context.Context, []byte) error)
	}
	c.handlers[queue] = handler
	return nil
}
