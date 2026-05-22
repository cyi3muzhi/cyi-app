package rmq

import (
	"time"

	"github.com/streadway/amqp"
)

type QueueOptions struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type ExchangeOptions struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

type BindOptions struct {
	QueueName    string
	BindingKey   string
	ExchangeName string
	NoWait       bool
	Args         amqp.Table
}

type PublishOptions struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
	Headers    amqp.Table
	Priority   uint8
	Expiration string
}

type ConsumeOptions struct {
	QueueName    string
	ConsumerTag  string
	AutoAck      bool
	Exclusive    bool
	NoLocal      bool
	NoWait       bool
	Args         amqp.Table
}

type Message struct {
	Body        []byte
	ContentType string
	Headers     amqp.Table
	Priority    uint8
	Expiration  string
}

func DefaultQueueOptions(name string) QueueOptions {
	return QueueOptions{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
}

func DefaultExchangeOptions(name, typ string) ExchangeOptions {
	return ExchangeOptions{
		Name:       name,
		Type:       typ,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
}

func DefaultPublishOptions(exchange, routingKey string) PublishOptions {
	return PublishOptions{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Mandatory:  false,
		Immediate:  false,
		Headers:    nil,
		Priority:   0,
		Expiration: "",
	}
}

func DefaultConsumeOptions(queueName string) ConsumeOptions {
	return ConsumeOptions{
		QueueName:   queueName,
		ConsumerTag: "",
		AutoAck:     true,
		Exclusive:   false,
		NoLocal:     false,
		NoWait:      false,
		Args:        nil,
	}
}

func NewDelayQueueOptions(name string, ttl time.Duration, dlxExchange, dlxRoutingKey string) QueueOptions {
	return QueueOptions{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args: amqp.Table{
			"x-message-ttl":             int64(ttl / time.Millisecond),
			"x-dead-letter-exchange":    dlxExchange,
			"x-dead-letter-routing-key": dlxRoutingKey,
		},
	}
}

func NewPriorityQueueOptions(name string, maxPriority uint8) QueueOptions {
	return QueueOptions{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args: amqp.Table{
			"x-max-priority": maxPriority,
		},
	}
}