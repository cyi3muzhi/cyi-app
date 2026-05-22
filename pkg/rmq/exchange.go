package rmq

import (
	"github.com/streadway/amqp"
)

type ExchangeManager struct {
	conn *Connection
}

func NewExchangeManager(conn *Connection) *ExchangeManager {
	return &ExchangeManager{conn: conn}
}

func (em *ExchangeManager) DeclareExchange(opts ExchangeOptions) error {
	return em.conn.Channel().ExchangeDeclare(
		opts.Name,
		opts.Type,
		opts.Durable,
		opts.AutoDelete,
		opts.Internal,
		opts.NoWait,
		opts.Args,
	)
}

func (em *ExchangeManager) DeleteExchange(name string, ifUnused bool) error {
	return em.conn.Channel().ExchangeDelete(name, ifUnused, false)
}

func (em *ExchangeManager) BindExchange(destination, routingKey, source string, noWait bool, args amqp.Table) error {
	return em.conn.Channel().ExchangeBind(destination, routingKey, source, noWait, args)
}