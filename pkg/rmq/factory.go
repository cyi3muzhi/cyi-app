package rmq

import (
	"time"

	"github.com/streadway/amqp"
)

type QueueFactory struct {
	conn *Connection
}

func NewQueueFactory(conn *Connection) *QueueFactory {
	return &QueueFactory{conn: conn}
}

func (f *QueueFactory) CreateSimpleQueue(name string) (*QueueInfo, error) {
	queueMgr := NewQueueManager(f.conn)
	
	opts := DefaultQueueOptions(name)
	q, err := queueMgr.DeclareQueue(opts)
	if err != nil {
		return nil, err
	}

	return &QueueInfo{
		Name:       q.Name,
		QueueMgr:   queueMgr,
		ExchangeMgr: NewExchangeManager(f.conn),
		Conn:       f.conn,
	}, nil
}

func (f *QueueFactory) CreateTopicQueue(exchangeName, queueName, bindingKey string) (*QueueInfo, error) {
	queueMgr := NewQueueManager(f.conn)
	exchangeMgr := NewExchangeManager(f.conn)

	exchangeOpts := DefaultExchangeOptions(exchangeName, "topic")
	if err := exchangeMgr.DeclareExchange(exchangeOpts); err != nil {
		return nil, err
	}

	queueOpts := DefaultQueueOptions(queueName)
	q, err := queueMgr.DeclareQueue(queueOpts)
	if err != nil {
		return nil, err
	}

	bindOpts := BindOptions{
		QueueName:    q.Name,
		BindingKey:   bindingKey,
		ExchangeName: exchangeName,
	}
	if err := queueMgr.BindQueue(bindOpts); err != nil {
		return nil, err
	}

	return &QueueInfo{
		Name:         q.Name,
		ExchangeName: exchangeName,
		BindingKey:   bindingKey,
		QueueMgr:     queueMgr,
		ExchangeMgr:  exchangeMgr,
		Conn:         f.conn,
	}, nil
}

func (f *QueueFactory) CreateFanoutQueue(exchangeName, queueName string) (*QueueInfo, error) {
	queueMgr := NewQueueManager(f.conn)
	exchangeMgr := NewExchangeManager(f.conn)

	exchangeOpts := DefaultExchangeOptions(exchangeName, "fanout")
	if err := exchangeMgr.DeclareExchange(exchangeOpts); err != nil {
		return nil, err
	}

	queueOpts := DefaultQueueOptions(queueName)
	q, err := queueMgr.DeclareQueue(queueOpts)
	if err != nil {
		return nil, err
	}

	bindOpts := BindOptions{
		QueueName:    q.Name,
		BindingKey:   "",
		ExchangeName: exchangeName,
	}
	if err := queueMgr.BindQueue(bindOpts); err != nil {
		return nil, err
	}

	return &QueueInfo{
		Name:         q.Name,
		ExchangeName: exchangeName,
		QueueMgr:     queueMgr,
		ExchangeMgr:  exchangeMgr,
		Conn:         f.conn,
	}, nil
}

func (f *QueueFactory) CreateDelayQueue(delayQueueName, targetQueueName string, ttl time.Duration) (*QueueInfo, error) {
	queueMgr := NewQueueManager(f.conn)
	exchangeMgr := NewExchangeManager(f.conn)

	targetExchangeName := "delay_target_" + delayQueueName
	
	exchangeOpts := DefaultExchangeOptions(targetExchangeName, "direct")
	if err := exchangeMgr.DeclareExchange(exchangeOpts); err != nil {
		return nil, err
	}

	targetQueueOpts := DefaultQueueOptions(targetQueueName)
	targetQueue, err := queueMgr.DeclareQueue(targetQueueOpts)
	if err != nil {
		return nil, err
	}

	bindOpts := BindOptions{
		QueueName:    targetQueue.Name,
		BindingKey:   "delay_key",
		ExchangeName: targetExchangeName,
	}
	if err := queueMgr.BindQueue(bindOpts); err != nil {
		return nil, err
	}

	delayQueueOpts := NewDelayQueueOptions(delayQueueName, ttl, targetExchangeName, "delay_key")
	_, err = queueMgr.DeclareQueue(delayQueueOpts)
	if err != nil {
		return nil, err
	}

	return &QueueInfo{
		Name:         targetQueue.Name,
		ExchangeName: targetExchangeName,
		DelayQueue:   delayQueueName,
		QueueMgr:     queueMgr,
		ExchangeMgr:  exchangeMgr,
		Conn:         f.conn,
	}, nil
}

func (f *QueueFactory) CreatePriorityQueue(name string, maxPriority uint8) (*QueueInfo, error) {
	queueMgr := NewQueueManager(f.conn)

	queueOpts := NewPriorityQueueOptions(name, maxPriority)
	q, err := queueMgr.DeclareQueue(queueOpts)
	if err != nil {
		return nil, err
	}

	return &QueueInfo{
		Name:       q.Name,
		QueueMgr:   queueMgr,
		ExchangeMgr: NewExchangeManager(f.conn),
		Conn:       f.conn,
	}, nil
}

func (f *QueueFactory) CreateDLXQueue(normalQueueName, dlxQueueName string, ttl time.Duration) (*QueueInfo, error) {
	queueMgr := NewQueueManager(f.conn)
	exchangeMgr := NewExchangeManager(f.conn)

	dlxExchangeName := "dlx_exchange_" + normalQueueName

	exchangeOpts := DefaultExchangeOptions(dlxExchangeName, "direct")
	if err := exchangeMgr.DeclareExchange(exchangeOpts); err != nil {
		return nil, err
	}

	dlxQueueOpts := DefaultQueueOptions(dlxQueueName)
	dlxQueue, err := queueMgr.DeclareQueue(dlxQueueOpts)
	if err != nil {
		return nil, err
	}

	bindOpts := BindOptions{
		QueueName:    dlxQueue.Name,
		BindingKey:   "dead_letter",
		ExchangeName: dlxExchangeName,
	}
	if err := queueMgr.BindQueue(bindOpts); err != nil {
		return nil, err
	}

	normalQueueOpts := NewDelayQueueOptions(normalQueueName, ttl, dlxExchangeName, "dead_letter")
	normalQueue, err := queueMgr.DeclareQueue(normalQueueOpts)
	if err != nil {
		return nil, err
	}

	return &QueueInfo{
		Name:         normalQueue.Name,
		ExchangeName: dlxExchangeName,
		DLXQueue:     dlxQueue.Name,
		QueueMgr:     queueMgr,
		ExchangeMgr:  exchangeMgr,
		Conn:         f.conn,
	}, nil
}

type QueueInfo struct {
	Name         string
	ExchangeName string
	BindingKey   string
	DelayQueue   string
	DLXQueue     string
	QueueMgr     *QueueManager
	ExchangeMgr  *ExchangeManager
	Conn         *Connection
}

func (qi *QueueInfo) NewProducer() *Producer {
	return NewProducer(qi.Conn)
}

func (qi *QueueInfo) NewConsumer() *Consumer {
	return NewConsumer(qi.Conn)
}

func (qi *QueueInfo) NewConfirmProducer() (*ConfirmProducer, error) {
	return NewConfirmProducer(qi.Conn)
}

func (qi *QueueInfo) Publish(body []byte, contentType string) error {
	producer := qi.NewProducer()
	
	var opts PublishOptions
	if qi.ExchangeName != "" {
		opts = DefaultPublishOptions(qi.ExchangeName, qi.BindingKey)
	} else {
		opts = DefaultPublishOptions("", qi.Name)
	}
	
	msg := Message{
		Body:        body,
		ContentType: contentType,
	}
	return producer.Publish(msg, opts)
}

func (qi *QueueInfo) PublishToDelay(body []byte, contentType string) error {
	if qi.DelayQueue == "" {
		return qi.Publish(body, contentType)
	}
	
	producer := qi.NewProducer()
	opts := DefaultPublishOptions("", qi.DelayQueue)
	
	msg := Message{
		Body:        body,
		ContentType: contentType,
	}
	return producer.Publish(msg, opts)
}

func (qi *QueueInfo) Consume(handler func(amqp.Delivery)) error {
	consumer := qi.NewConsumer()
	opts := DefaultConsumeOptions(qi.Name)
	
	msgs, err := consumer.Consume(opts)
	if err != nil {
		return err
	}
	
	go func() {
		for d := range msgs {
			handler(d)
		}
	}()
	
	return nil
}