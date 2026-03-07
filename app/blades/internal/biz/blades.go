package biz

import (
	"context"

	"cyi-app/app/blades/internal/conf"
	pkgblades "cyi-app/pkg/blades"

	"github.com/go-kratos/blades"
)

// BladesUseCase 处理Blades相关的业务逻辑
type BladesUseCase struct {
	agent blades.Agent
	repo  BladesRepo
}

type BladesRepo interface {
	Chat(ctx context.Context, message string) (string, error)
}

// NewBladesUseCase 创建BladesUseCase实例
func NewBladesUseCase(repo BladesRepo, c *conf.Blades) (*BladesUseCase, error) {
	agent, err := pkgblades.CreateAgent(c.ApiKey)
	if err != nil {
		return nil, err
	}
	return &BladesUseCase{
		repo:  repo,
		agent: agent,
	}, nil
}

// Chat 处理聊天请求
func (uc *BladesUseCase) Chat(ctx context.Context, message string) (string, error) {
	// 记录请求
	_, err := uc.repo.Chat(ctx, message)
	if err != nil {
		// 即使 repo 调用失败，也继续处理请求
		// return "", err
	}
	
	// 尝试调用外部 API
	response, err := pkgblades.RunAgent(ctx, uc.agent, message)
	if err != nil {
		// 如果 API 调用失败，返回错误信息但不中断服务
		return "Error: " + err.Error(), nil
	}
	
	return response, nil
}
