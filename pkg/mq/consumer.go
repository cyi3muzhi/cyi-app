package mq

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/streadway/amqp"
)

// Consumer 消息消费者
type Consumer struct {
	router *MessageRouter
	mq     *RabbitMQ
	queue  string
	stopCh chan struct{}
	wg     sync.WaitGroup
	log    *log.Helper
}

// NewConsumer 创建消息消费者
func NewConsumer(router *MessageRouter, mq *RabbitMQ, queue string, logger log.Logger) *Consumer {
	return &Consumer{
		router: router,
		mq:     mq,
		queue:  queue,
		stopCh: make(chan struct{}),
		log:    log.NewHelper(logger),
	}
}

// Start 启动消费者
func (c *Consumer) Start(ctx context.Context) error {
	// 声明队列
	_, err := c.mq.DeclareQueue(c.queue, true, false, false, false, nil)
	if err != nil {
		c.log.Error("Failed to declare queue: %v", err)
		return err
	}

	// 开始消费
	msgs, err := c.mq.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		c.log.Error("Failed to start consuming: %v", err)
		return err
	}

	c.wg.Add(1)
	go c.consumeLoop(ctx, msgs)

	c.log.Info("Started consumer for queue: %s", c.queue)
	return nil
}

// Stop 停止消费者
func (c *Consumer) Stop() {
	close(c.stopCh)
	c.wg.Wait()
	c.log.Info("Stopped consumer for queue: %s", c.queue)
}

// consumeLoop 消费循环
func (c *Consumer) consumeLoop(ctx context.Context, msgs <-chan amqp.Delivery) {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}

			c.processMessage(ctx, &msg)
		}
	}
}

// processMessage 处理消息
func (c *Consumer) processMessage(ctx context.Context, msg *amqp.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			c.log.Error("Failed to ack message: %v", err)
		}
	}()

	// 解析消息头，获取消息类型
	messageType, ok := msg.Headers["message_type"].(string)
	if !ok || messageType == "" {
		c.log.Error("Message type not set")
		return
	}

	// 处理消息
	if err := c.router.HandleMessage(ctx, messageType, msg.Body); err != nil {
		c.log.Error("Failed to handle message: %v", err)
	}
}
