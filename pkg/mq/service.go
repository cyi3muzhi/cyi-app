package mq

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/streadway/amqp"
)

// Service 消息队列服务
type Service struct {
	rmq    *RabbitMQ
	router *MessageRouter
	opts   *Options
}

// NewService 创建消息队列服务
func NewService(opts ...Option) (*Service, error) {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	// 验证配置
	if options.Config == nil {
		return nil, fmt.Errorf("mq config is required")
	}
	if options.Logger == nil {
		options.Logger = log.NewStdLogger(os.Stdout)
	}

	// 使用配置中的队列列表，如果没有指定则使用默认队列
	if len(options.Queues) == 0 && len(options.Config.Queues) > 0 {
		options.Queues = options.Config.Queues
	}

	// 创建组件
	rmq := NewRabbitMQ(options.Config, options.Logger)
	router := NewMessageRouter(options.Logger)

	return &Service{
		rmq:    rmq,
		router: router,
		opts:   options,
	}, nil
}

// RegisterHandler 注册消息处理器
func (s *Service) RegisterHandler(messageType string, handler MessageHandler) {
	s.router.RegisterHandler(messageType, handler)
}

// Start 启动消息队列服务
func (s *Service) Start(ctx context.Context) error {
	// 连接到RabbitMQ
	if err := s.rmq.Connect(); err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 启动心跳检测
	go s.rmq.StartHeartbeat(ctx)

	// 启动队列消费者
	for _, queue := range s.opts.Queues {
		consumer := NewConsumer(s.router, s.rmq, queue, s.opts.Logger)
		if err := consumer.Start(ctx); err != nil {
			log.NewHelper(s.opts.Logger).Error("Failed to start consumer for queue %s: %v", queue, err)
			continue
		}
	}

	log.NewHelper(s.opts.Logger).Info("Message queue service started successfully")
	return nil
}

// Close 关闭消息队列服务
func (s *Service) Close() error {
	return s.rmq.Close()
}

// Publish 发布消息
func (s *Service) Publish(exchange, routingKey string, mandatory, immediate bool, msg amqp.Publishing) error {
	return s.rmq.Publish(exchange, routingKey, mandatory, immediate, msg)
}
