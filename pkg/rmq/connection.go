package rmq

import (
	"sync"

	"github.com/streadway/amqp"
)

type Connection struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	mu        sync.Mutex
	url       string
	isClosed  bool
}

func NewConnection(url string) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Connection{
		conn:      conn,
		channel:   ch,
		url:       url,
		isClosed:  false,
	}, nil
}

func (c *Connection) Channel() *amqp.Channel {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.channel
}

func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return nil
	}

	c.isClosed = true
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}

func (c *Connection) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isClosed {
		c.channel.Close()
		c.conn.Close()
	}

	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	c.conn = conn
	c.channel = ch
	c.isClosed = false
	return nil
}

func (c *Connection) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	return c.conn.NotifyClose(receiver)
}