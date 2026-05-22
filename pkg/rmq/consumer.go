package rmq

import (
	"github.com/streadway/amqp"
)

type Consumer struct {
	conn *Connection
}

func NewConsumer(conn *Connection) *Consumer {
	return &Consumer{conn: conn}
}

func (c *Consumer) Consume(opts ConsumeOptions) (<-chan amqp.Delivery, error) {
	return c.conn.Channel().Consume(
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