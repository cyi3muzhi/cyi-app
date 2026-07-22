package rmq

import (
	"github.com/streadway/amqp"
)

// ExchangeManager 交换机管理，每个操作使用独立临时 Channel
type ExchangeManager struct {
	conn *Connection
}

func NewExchangeManager(conn *Connection) *ExchangeManager {
	return &ExchangeManager{conn: conn}
}

func (em *ExchangeManager) withChannel(f func(ch *amqp.Channel) error) error {
	ch, err := em.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	return f(ch)
}

func (em *ExchangeManager) DeclareExchange(opts ExchangeOptions) error {
	return em.withChannel(func(ch *amqp.Channel) error {
		return ch.ExchangeDeclare(
			opts.Name, opts.Type,
			opts.Durable, opts.AutoDelete,
			opts.Internal, opts.NoWait, opts.Args,
		)
	})
}

func (em *ExchangeManager) DeleteExchange(name string, ifUnused bool) error {
	return em.withChannel(func(ch *amqp.Channel) error {
		return ch.ExchangeDelete(name, ifUnused, false)
	})
}

func (em *ExchangeManager) BindExchange(destination, routingKey, source string, noWait bool, args amqp.Table) error {
	return em.withChannel(func(ch *amqp.Channel) error {
		return ch.ExchangeBind(destination, routingKey, source, noWait, args)
	})
}