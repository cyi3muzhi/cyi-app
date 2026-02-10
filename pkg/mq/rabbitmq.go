package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/streadway/amqp"
)

// RabbitMQ RabbitMQ连接管理
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *MessageQueue_RabbitMQ
	log     *log.Helper
}

// NewRabbitMQ 创建RabbitMQ连接管理实例
func NewRabbitMQ(cfg *MessageQueue, logger log.Logger) *RabbitMQ {
	return &RabbitMQ{
		config: cfg.Rabbitmq,
		log:    log.NewHelper(logger),
	}
}

// Connect 连接到RabbitMQ
func (r *RabbitMQ) Connect() error {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		r.config.Username,
		r.config.Password,
		r.config.Host,
		r.config.Port,
		r.config.VirtualHost,
	)

	var err error
	r.conn, err = amqp.Dial(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	r.log.Info("Connected to RabbitMQ")
	return nil
}

// Reconnect 重新连接到RabbitMQ
func (r *RabbitMQ) Reconnect() error {
	r.log.Info("Reconnecting to RabbitMQ...")
	return r.Connect()
}

// Close 关闭RabbitMQ连接
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	r.log.Info("Closed RabbitMQ connection")
	return nil
}

// Channel 获取RabbitMQ通道
func (r *RabbitMQ) Channel() *amqp.Channel {
	return r.channel
}

// IsConnected 检查是否连接到RabbitMQ
func (r *RabbitMQ) IsConnected() bool {
	if r.conn == nil {
		return false
	}
	return !r.conn.IsClosed()
}

// DeclareQueue 声明队列
func (r *RabbitMQ) DeclareQueue(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	if !r.IsConnected() {
		if err := r.Reconnect(); err != nil {
			return amqp.Queue{}, err
		}
	}

	return r.channel.QueueDeclare(
		name,       // 队列名
		durable,    // 持久化
		autoDelete, // 自动删除
		exclusive,  // 排他性
		noWait,     // 不等待
		args,       // 参数
	)
}

// DeclareExchange 声明交换机
func (r *RabbitMQ) DeclareExchange(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	if !r.IsConnected() {
		if err := r.Reconnect(); err != nil {
			return err
		}
	}

	return r.channel.ExchangeDeclare(
		name,       // 交换机名
		kind,       // 类型
		durable,    // 持久化
		autoDelete, // 自动删除
		internal,   // 内部
		noWait,     // 不等待
		args,       // 参数
	)
}

// BindQueue 绑定队列到交换机
func (r *RabbitMQ) BindQueue(queueName, exchange, routingKey string, noWait bool, args amqp.Table) error {
	if !r.IsConnected() {
		if err := r.Reconnect(); err != nil {
			return err
		}
	}

	return r.channel.QueueBind(
		queueName,  // 队列名
		routingKey, // 路由键
		exchange,   // 交换机
		noWait,     // 不等待
		args,       // 参数
	)
}

// Publish 发布消息
func (r *RabbitMQ) Publish(exchange, routingKey string, mandatory, immediate bool, msg amqp.Publishing) error {
	if !r.IsConnected() {
		if err := r.Reconnect(); err != nil {
			return err
		}
	}

	return r.channel.Publish(
		exchange,   // 交换机
		routingKey, // 路由键
		mandatory,  // 强制
		immediate,  // 立即
		msg,        // 消息
	)
}

// Consume 消费消息
func (r *RabbitMQ) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if !r.IsConnected() {
		if err := r.Reconnect(); err != nil {
			return nil, err
		}
	}

	return r.channel.Consume(
		queue,     // 队列名
		consumer,  // 消费者
		autoAck,   // 自动确认
		exclusive, // 排他性
		noLocal,   // 非本地
		noWait,    // 不等待
		args,      // 参数
	)
}

// StartHeartbeat 启动心跳检测
func (r *RabbitMQ) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(r.config.Heartbeat) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !r.IsConnected() {
				if err := r.Reconnect(); err != nil {
					r.log.Error("Heartbeat reconnect failed: %v", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
