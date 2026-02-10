package mq

import (
	"github.com/go-kratos/kratos/v2/log"
)

// MessageQueue_RabbitMQ RabbitMQ配置
type MessageQueue_RabbitMQ struct {
	Host              string
	Port              int32
	Username          string
	Password          string
	VirtualHost       string
	ConnectionTimeout int32
	Heartbeat         int32
}

// MessageQueue 消息队列配置
type MessageQueue struct {
	Rabbitmq *MessageQueue_RabbitMQ
	Queues   []string
}

// Options 消息队列服务配置选项
type Options struct {
	Config *MessageQueue
	Logger log.Logger
	Queues []string
}

// Option 配置选项函数
type Option func(*Options)

// WithConfig 设置消息队列配置
func WithConfig(config *MessageQueue) Option {
	return func(o *Options) {
		o.Config = config
	}
}

// WithLogger 设置日志器
func WithLogger(logger log.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithQueues 设置要监听的队列
func WithQueues(queues []string) Option {
	return func(o *Options) {
		o.Queues = queues
	}
}
