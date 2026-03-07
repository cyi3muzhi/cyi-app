package blades

import (
	"context"
	"fmt"
)

// DeepSeekDemo 运行DeepSeek模型的演示
func DeepSeekDemo() error {
	// 使用提供的DeepSeek API密钥
	apiKey := "sk-9563c2ca6c3e44cd9c51b8faf5dd82e3"
	if apiKey == "" {
		return fmt.Errorf("DeepSeek API key not provided")
	}

	// 创建DeepSeek Agent
	ag, err := CreateAgent(apiKey)
	if err != nil {
		return fmt.Errorf("failed to create deepseek agent: %w", err)
	}

	// 测试问题
	questions := []string{
		"2 + 2 等于多少？",
		"讲一个笑话",
		"法国的首都是什么？",
		"100 + 200 等于多少？",
	}

	ctx := context.Background()

	// 运行测试
	fmt.Println("\n=== DeepSeek Model Demo ===")
	for _, question := range questions {
		fmt.Printf("\nQuestion: %s\n", question)
		response, err := RunAgent(ctx, ag, question)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("Response: %s\n", response)
	}

	return nil
}
