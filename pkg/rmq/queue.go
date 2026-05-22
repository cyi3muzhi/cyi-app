package rmq

import (
	"github.com/streadway/amqp"
)

type QueueManager struct {
	conn *Connection
}

func NewQueueManager(conn *Connection) *QueueManager {
	return &QueueManager{conn: conn}
}

func (qm *QueueManager) DeclareQueue(opts QueueOptions) (amqp.Queue, error) {
	return qm.conn.Channel().QueueDeclare(
		opts.Name,
		opts.Durable,
		opts.AutoDelete,
		opts.Exclusive,
		opts.NoWait,
		opts.Args,
	)
}

func (qm *QueueManager) BindQueue(opts BindOptions) error {
	return qm.conn.Channel().QueueBind(
		opts.QueueName,
		opts.BindingKey,
		opts.ExchangeName,
		opts.NoWait,
		opts.Args,
	)
}

func (qm *QueueManager) DeleteQueue(name string, ifUnused, ifEmpty bool) error {
	_, err := qm.conn.Channel().QueueDelete(name, ifUnused, ifEmpty, false)
	return err
}

func (qm *QueueManager) PurgeQueue(name string) (int, error) {
	return qm.conn.Channel().QueuePurge(name, false)
}