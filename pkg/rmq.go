package main

import (
	"cyi-app/pkg/rmq"
	"fmt"
	"log"
	"time"
)

func main() {
	conn, err := rmq.NewConnection("amqp://admin:admin@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	factory := rmq.NewQueueFactory(conn)

	fmt.Println("=== 1. 使用简单队列 ===")
	simpleQueue, err := factory.CreateSimpleQueue(rmq.QueueOrderCreated)
	if err != nil {
		log.Fatalf("Failed to create simple queue: %v", err)
	}
	err = simpleQueue.Publish([]byte("Hello Simple Queue"), rmq.ContentTypeText)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Println("Message published to simple queue")

	fmt.Println("\n=== 2. 使用主题队列 ===")
	topicQueue, err := factory.CreateTopicQueue(rmq.ExchangeLogs, rmq.QueueLogsAll, "order.log")
	if err != nil {
		log.Fatalf("Failed to create topic queue: %v", err)
	}
	err = topicQueue.Publish([]byte("Hello Topic Queue"), rmq.ContentTypeText)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Println("Message published to topic queue")

	fmt.Println("\n=== 3. 使用扇出队列 (Fanout) ===")
	// 使用独立的交换机，避免与 topic 交换机冲突q	
	fanoutQueue, err := factory.CreateFanoutQueue(rmq.ExchangeProduct, rmq.QueueDLXOrder)
	if err != nil {
		log.Fatalf("Failed to create fanout queue: %v", err)
	}
	err = fanoutQueue.Publish([]byte("Hello Fanout Queue - Broadcast to all"), rmq.ContentTypeText)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Println("Message published to fanout queue (broadcast)")

	fmt.Println("\n=== 4. 使用延迟队列 (Delay) ===")
	delayQueue, err := factory.CreateDelayQueue("order_delay_queue", "order_target_queue", 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to create delay queue: %v", err)
	}
	err = delayQueue.PublishToDelay([]byte("Hello Delay Queue - Will be delivered after 5 seconds"), rmq.ContentTypeText)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Println("Message published to delay queue (will be delivered after 5 seconds)")

	fmt.Println("\n=== 5. 使用优先级队列 (Priority) ===")
	priorityQueue, err := factory.CreatePriorityQueue(rmq.QueuePriorityOrder, rmq.DefaultMaxPriority)
	if err != nil {
		log.Fatalf("Failed to create priority queue: %v", err)
	}
	// 发布高优先级消息
	err = priorityQueue.Publish([]byte("Hello Priority Queue - High Priority"), rmq.ContentTypeText)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Println("High priority message published to priority queue")

	fmt.Println("\n=== 6. 使用死信队列 (DLX) ===")
	dlxQueue, err := factory.CreateDLXQueue(rmq.QueueNormalOrder, rmq.QueueDLXOrder, 30*time.Second)
	if err != nil {
		log.Fatalf("Failed to create DLX queue: %v", err)
	}
	err = dlxQueue.Publish([]byte("Hello DLX Queue - Will go to DLX if not consumed"), rmq.ContentTypeText)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Println("Message published to DLX queue")

	fmt.Println("\n=== All messages published successfully ===")
}
