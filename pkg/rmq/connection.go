package rmq

import (
	"sync"

	"github.com/streadway/amqp"
)

// Connection 封装 AMQP 连接，每个 Consumer/Producer 应使用独立的 Channel
type Connection struct {
	conn     *amqp.Connection
	mu       sync.RWMutex
	url      string
	isClosed bool
}

// NewConnection 创建 AMQP 连接
func NewConnection(url string) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return &Connection{
		conn:     conn,
		url:      url,
		isClosed: false,
	}, nil
}

// Channel 创建一个新的 AMQP Channel（每个 goroutine 应使用独立 Channel）
func (c *Connection) Channel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isClosed {
		return nil, amqp.ErrClosed
	}
	return c.conn.Channel()
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return nil
	}

	c.isClosed = true
	return c.conn.Close()
}

// Reconnect 重新连接（先尽力关闭旧连接，再拨号新连接）
func (c *Connection) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 尽力关闭旧连接（忽略已关闭的错误）
	if !c.isClosed {
		_ = c.conn.Close()
	}

	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}

	c.conn = conn
	c.isClosed = false
	return nil
}

// NotifyClose 注册连接关闭通知
func (c *Connection) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn.NotifyClose(receiver)
}