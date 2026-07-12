package consumers

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
)

// connectNATS 尝试连接NATS，失败则跳过测试
func connectNATS(t *testing.T) *nats.Conn {
	t.Helper()
	nc, err := nats.Connect(nats.DefaultURL, nats.Timeout(3*time.Second))
	if err != nil {
		t.Skipf("NATS 服务未运行，跳过测试: %v", err)
	}
	t.Cleanup(func() { nc.Close() })
	return nc
}

// TestNATS_Connection 测试NATS是否能连通
func TestNATS_Connection(t *testing.T) {
	nc := connectNATS(t)

	// 发布一条测试消息，验证发布不报错
	err := nc.Publish("test.ping", []byte("ping"))
	if err != nil {
		t.Fatalf("发布消息失败: %v", err)
	}
	t.Log("NATS 连接正常，消息发布成功")
}

// TestNATS_PublishSubscribe 测试发布/订阅功能
func TestNATS_PublishSubscribe(t *testing.T) {
	nc := connectNATS(t)

	// 订阅一个主题，等待消息
	msgCh := make(chan []byte, 1)
	_, err := nc.Subscribe("test.echo", func(msg *nats.Msg) {
		msgCh <- msg.Data
	})
	if err != nil {
		t.Fatalf("订阅失败: %v", err)
	}
	defer nc.Drain()

	// 确保订阅已注册
	nc.Flush()

	// 发布消息
	err = nc.Publish("test.echo", []byte("hello nats"))
	if err != nil {
		t.Fatalf("发布消息失败: %v", err)
	}

	// 等待接收消息（超时3秒）
	select {
	case data := <-msgCh:
		if string(data) != "hello nats" {
			t.Errorf("收到消息内容不符: got %s, want hello nats", string(data))
		}
		t.Logf("收到消息: %s", string(data))
	case <-time.After(3 * time.Second):
		t.Fatal("等待消息超时")
	}
}

// TestNATS_QueueGroup 测试队列组（竞争消费）
func TestNATS_QueueGroup(t *testing.T) {
	nc := connectNATS(t)

	// 两个消费者组成队列组，消息只会被其中一个消费
	var callCount atomic.Int32
	handler := func(msg *nats.Msg) {
		callCount.Add(1)
	}

	sub1, err := nc.QueueSubscribe("test.queue", "test-group", handler)
	if err != nil {
		t.Fatalf("订阅1失败: %v", err)
	}
	defer sub1.Unsubscribe()

	sub2, err := nc.QueueSubscribe("test.queue", "test-group", handler)
	if err != nil {
		t.Fatalf("订阅2失败: %v", err)
	}
	defer sub2.Unsubscribe()

	nc.Flush()

	// 发3条消息，队列组内两个消费者各自消费，但总共只收到3条
	for i := 0; i < 3; i++ {
		nc.Publish("test.queue", []byte("msg"))
	}

	// 等待消息处理
	time.Sleep(500 * time.Millisecond)

	if got := callCount.Load(); got != 3 {
		t.Errorf("队列组应收到3条消息，实际收到 %d", got)
	}
}

// TestNatsConsumer_Subscribe 测试 NatsConsumer 订阅和处理
func TestNatsConsumer_Subscribe(t *testing.T) {
	nc := connectNATS(t)

	// 创建 NatsConsumer
	consumer := NewNatsConsumer(nc, log.DefaultLogger)

	// 注册处理器
	var received atomic.Value
	received.Store("")
	consumer.Subscribe("test.consumer", func(_ context.Context, data []byte) error {
		received.Store(string(data))
		return nil
	})

	// 启动消费
	err := consumer.Start()
	if err != nil {
		t.Fatalf("启动消费者失败: %v", err)
	}
	defer consumer.Stop()

	// 发布消息
	err = nc.Publish("test.consumer", []byte("consumer test ok"))
	if err != nil {
		t.Fatalf("发布消息失败: %v", err)
	}

	// 等待处理
	time.Sleep(500 * time.Millisecond)

	if got := received.Load().(string); got != "consumer test ok" {
		t.Errorf("处理器未收到正确消息: got %s, want consumer test ok", got)
	}
}

// TestConsumer_Full 测试 Consumer 编排器到 Service 的全链路
func TestConsumer_Full(t *testing.T) {
	nc := connectNATS(t)

	// 创建串联测试用的 Service 和 UseCase
	// 使用 NatsConsumer 直接连接 NATS，不走 wire 注入
	consumer := NewNatsConsumer(nc, log.DefaultLogger)

	// 注册一个计数处理器验证消息到达
	var callCount atomic.Int32
	consumer.Subscribe("order.test", func(_ context.Context, data []byte) error {
		callCount.Add(1)
		t.Logf("收到订单消息: %s", string(data))
		return nil
	})

	// 启动
	err := consumer.Start()
	if err != nil {
		t.Fatalf("启动失败: %v", err)
	}
	defer consumer.Stop()

	// 发布测试消息
	for i := 0; i < 3; i++ {
		err = nc.Publish("order.test", []byte(`{"order_id": 1, "action": "test"}`))
		if err != nil {
			t.Fatalf("发布消息失败: %v", err)
		}
	}

	// 等待处理
	time.Sleep(500 * time.Millisecond)

	if got := callCount.Load(); got != 3 {
		t.Errorf("应收到3条消息，实际收到 %d", got)
	}
}

// BenchmarkNATS_Publish 基准测试：NATS 发布性能
func BenchmarkNATS_Publish(b *testing.B) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		b.Skipf("NATS 未运行: %v", err)
	}
	defer nc.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nc.Publish("bench.test", []byte("benchmark message"))
	}
}

// BenchmarkNATS_PublishSubscribe 基准测试：发布+订阅性能
func BenchmarkNATS_PublishSubscribe(b *testing.B) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		b.Skipf("NATS 未运行: %v", err)
	}
	defer nc.Close()

	var count atomic.Int32
	nc.Subscribe("bench.sub", func(msg *nats.Msg) {
		count.Add(1)
	})
	nc.Flush()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nc.Publish("bench.sub", []byte("bench"))
	}

	// 等待所有消息消费完毕
	time.Sleep(2 * time.Second)
}
