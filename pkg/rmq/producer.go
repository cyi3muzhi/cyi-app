package rmq

import (
	"github.com/streadway/amqp"
)

type Producer struct {
	conn *Connection
}

func NewProducer(conn *Connection) *Producer {
	return &Producer{conn: conn}
}

func (p *Producer) Publish(msg Message, opts PublishOptions) error {
	publishing := amqp.Publishing{
		ContentType: msg.ContentType,
		Body:        msg.Body,
		Headers:     msg.Headers,
		Priority:    msg.Priority,
		Expiration:  msg.Expiration,
	}

	if opts.Headers != nil {
		publishing.Headers = opts.Headers
	}
	if opts.Priority > 0 {
		publishing.Priority = opts.Priority
	}
	if opts.Expiration != "" {
		publishing.Expiration = opts.Expiration
	}

	return p.conn.Channel().Publish(
		opts.Exchange,
		opts.RoutingKey,
		opts.Mandatory,
		opts.Immediate,
		publishing,
	)
}

func (p *Producer) PublishString(body, contentType string, opts PublishOptions) error {
	msg := Message{
		Body:        []byte(body),
		ContentType: contentType,
	}
	return p.Publish(msg, opts)
}

type ConfirmProducer struct {
	conn         *Connection
	confirmChan  chan amqp.Confirmation
}

func NewConfirmProducer(conn *Connection) (*ConfirmProducer, error) {
	err := conn.Channel().Confirm(false)
	if err != nil {
		return nil, err
	}

	return &ConfirmProducer{
		conn:        conn,
		confirmChan: conn.Channel().NotifyPublish(make(chan amqp.Confirmation, 1)),
	}, nil
}

func (cp *ConfirmProducer) PublishWithConfirm(msg Message, opts PublishOptions) (bool, error) {
	err := cp.conn.Channel().Publish(
		opts.Exchange,
		opts.RoutingKey,
		opts.Mandatory,
		opts.Immediate,
		amqp.Publishing{
			ContentType: msg.ContentType,
			Body:        msg.Body,
			Headers:     msg.Headers,
			Priority:    msg.Priority,
			Expiration:  msg.Expiration,
		},
	)
	if err != nil {
		return false, err
	}

	confirm := <-cp.confirmChan
	return confirm.Ack, nil
}