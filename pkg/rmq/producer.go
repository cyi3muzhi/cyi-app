package rmq

import (
	"github.com/streadway/amqp"
)

// Producer 消息生产者，持有独立 Channel（线程安全）
type Producer struct {
	conn    *Connection
	channel *amqp.Channel
}

// NewProducer 创建生产者并创建独立 Channel
func NewProducer(conn *Connection) (*Producer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &Producer{conn: conn, channel: ch}, nil
}

// Close 关闭生产者持有的 Channel
func (p *Producer) Close() error {
	return p.channel.Close()
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

	return p.channel.Publish(
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

// ConfirmProducer 可靠消息生产者（发布确认模式）
type ConfirmProducer struct {
	producer    *Producer
	confirmChan chan amqp.Confirmation
}

// NewConfirmProducer 创建确认模式生产者
func NewConfirmProducer(conn *Connection) (*ConfirmProducer, error) {
	p, err := NewProducer(conn)
	if err != nil {
		return nil, err
	}

	if err := p.channel.Confirm(false); err != nil {
		p.channel.Close()
		return nil, err
	}

	return &ConfirmProducer{
		producer:    p,
		confirmChan: p.channel.NotifyPublish(make(chan amqp.Confirmation, 1)),
	}, nil
}

// Close 关闭确认生产者
func (cp *ConfirmProducer) Close() error {
	return cp.producer.Close()
}

func (cp *ConfirmProducer) PublishWithConfirm(msg Message, opts PublishOptions) (bool, error) {
	err := cp.producer.channel.Publish(
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