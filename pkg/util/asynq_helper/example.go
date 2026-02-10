package asynqhelper

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// ExampleClient 展示如何使用客户端创建任务
func ExampleClient() {
	// 创建客户端
	client, err := NewClient("localhost:6379", "", 0)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	// 创建邮件任务
	emailPayload := &EmailTaskPayload{
		To:      "user@example.com",
		Subject: "Welcome to our service",
		Body:    "Hello! Welcome to our service. We're glad to have you here.",
	}

	info, err := client.CreateEmailTask(emailPayload)
	if err != nil {
		fmt.Printf("Failed to create email task: %v\n", err)
		return
	}

	fmt.Printf("Created email task with ID: %s\n", info.ID)

	// 创建短信任务
	smsPayload := &SMSTaskPayload{
		Phone:   "+1234567890",
		Content: "Your verification code is 123456",
	}

	info, err = client.CreateSMSTask(smsPayload)
	if err != nil {
		fmt.Printf("Failed to create SMS task: %v\n", err)
		return
	}

	fmt.Printf("Created SMS task with ID: %s\n", info.ID)

	// 创建延迟任务
	notificationPayload := &NotificationTaskPayload{
		UserID:  "123",
		Title:   "Reminder",
		Message: "Don't forget to complete your profile",
		Type:    "info",
	}

	info, err = client.ScheduleTask(TaskTypeNotification, notificationPayload, 10*time.Minute)
	if err != nil {
		fmt.Printf("Failed to schedule notification task: %v\n", err)
		return
	}

	fmt.Printf("Scheduled notification task with ID: %s\n", info.ID)
}

// ExampleServer 展示如何使用服务器处理任务
func ExampleServer() {
	// 创建服务器
	server, err := NewServer("localhost:6379", "", 0, 10)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	// 创建默认任务处理器
	handler := NewDefaultHandler()

	// 注册自定义任务处理器
	handler.RegisterHandlerFunc("custom:task", func(ctx context.Context, task *asynq.Task) error {
		fmt.Println("Processing custom task")
		return nil
	})

	// 启动服务器
	go func() {
		if err := server.Start(handler); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	}()

	// 运行一段时间后停止服务器
	time.Sleep(1 * time.Minute)

	// 优雅关闭服务器
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown()
}

// ExampleComplete 展示完整的使用示例，包括客户端和服务器
func ExampleComplete() {
	// 启动服务器（在后台）
	go func() {
		ExampleServer()
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 创建客户端并创建任务
	ExampleClient()

	// 等待任务处理
	time.Sleep(5 * time.Second)
}
