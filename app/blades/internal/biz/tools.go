package biz

// ToolRegistry 工具注册中心
type ToolRegistry struct {
	list []interface{}
}

// NewToolRegistry 创建工具注册中心
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{}
}

// GetTools 获取所有工具
func (r *ToolRegistry) GetTools() []interface{} {
	return r.list
}

// Add 添加工具
func (r *ToolRegistry) Add(tool interface{}) {
	r.list = append(r.list, tool)
}
