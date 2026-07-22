package rmq

import (
	"github.com/streadway/amqp"
)

// QueueManager 队列管理，每个操作使用独立临时 Channel
type QueueManager struct {
	conn *Connection
}

func NewQueueManager(conn *Connection) *QueueManager {
	return &QueueManager{conn: conn}
}

func (qm *QueueManager) withChannel(f func(ch *amqp.Channel) error) error {
	ch, err := qm.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	return f(ch)
}

func (qm *QueueManager) DeclareQueue(opts QueueOptions) (amqp.Queue, error) {
	var queue amqp.Queue
	err := qm.withChannel(func(ch *amqp.Channel) error {
		var err error
		queue, err = ch.QueueDeclare(
			opts.Name, opts.Durable,
			opts.AutoDelete, opts.Exclusive,
			opts.NoWait, opts.Args,
		)
		return err
	})
	return queue, err
}

func (qm *QueueManager) BindQueue(opts BindOptions) error {
	return qm.withChannel(func(ch *amqp.Channel) error {
		return ch.QueueBind(
			opts.QueueName, opts.BindingKey,
			opts.ExchangeName, opts.NoWait, opts.Args,
		)
	})
}

func (qm *QueueManager) DeleteQueue(name string, ifUnused, ifEmpty bool) error {
	return qm.withChannel(func(ch *amqp.Channel) error {
		_, err := ch.QueueDelete(name, ifUnused, ifEmpty, false)
		return err
	})
}

func (qm *QueueManager) PurgeQueue(name string) (int, error) {
	var count int
	err := qm.withChannel(func(ch *amqp.Channel) error {
		var err error
		count, err = ch.QueuePurge(name, false)
		return err
	})
	return count, err
}