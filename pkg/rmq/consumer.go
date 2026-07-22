package rmq

import (
	"github.com/streadway/amqp"
)

// Consumer 消息消费者，持有独立 Channel（线程安全）
type Consumer struct {
	conn    *Connection
	channel *amqp.Channel
}

// NewConsumer 创建消费者并创建独立 Channel
func NewConsumer(conn *Connection) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &Consumer{conn: conn, channel: ch}, nil
}

// Channel 返回消费者持有的 AMQP Channel
func (c *Consumer) Channel() *amqp.Channel {
	return c.channel
}

// Close 关闭消费者持有的 Channel
func (c *Consumer) Close() error {
	return c.channel.Close()
}

func (c *Consumer) Consume(opts ConsumeOptions) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		opts.QueueName,
		opts.ConsumerTag,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)
}

func (c *Consumer) ConsumeWithHandler(opts ConsumeOptions, handler func(amqp.Delivery) error) error {
	msgs, err := c.Consume(opts)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			if err := handler(d); err != nil {
				if !opts.AutoAck {
					d.Nack(false, true)
				}
			} else if !opts.AutoAck {
				d.Ack(false)
			}
		}
	}()

	return nil
}