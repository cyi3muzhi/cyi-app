package biz

import "github.com/go-kratos/blades/tools"

// ToolRegistry 工具注册中心
type ToolRegistry struct {
	list []tools.Tool
}

// NewToolRegistry 创建工具注册中心
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{}
}

// GetTools 获取所有工具
func (r *ToolRegistry) GetTools() []tools.Tool {
	return r.list
}

// Add 添加工具
func (r *ToolRegistry) Add(tool tools.Tool) {
	r.list = append(r.list, tool)
}
