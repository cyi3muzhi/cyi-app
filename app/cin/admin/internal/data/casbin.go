package data

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	mid "cyi-app/pkg/kratos/server/middleware"
)

// NewCasbinEnforcer 创建 Casbin Enforcer 实例（自动从 configs 目录加载模型和策略）
func NewCasbinEnforcer() (*casbin.Enforcer, error) {
	configDir := GetConfigDir()
	modelPath := filepath.Join(configDir, "rbac_model.conf")
	policyPath := filepath.Join(configDir, "rbac_policy.csv")

	// 从文件加载模型
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("加载 Casbin 模型失败: %w", err)
	}

	// 从文件加载策略
	a := fileadapter.NewAdapter(policyPath)

	// 创建 Enforcer
	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, fmt.Errorf("创建 Casbin Enforcer 失败: %w", err)
	}

	// 启用自动保存（策略修改时自动写回文件）
	e.EnableAutoSave(true)

	return e, nil
}

// ProvideEnforcerProvider 将 *casbin.Enforcer 转换为 EnforcerProvider 接口（用于 wire 注入）
func ProvideEnforcerProvider(e *casbin.Enforcer) mid.EnforcerProvider {
	return e
}

// GetConfigDir 获取配置文件目录（用于定位 rbac_model.conf 和 rbac_policy.csv）
func GetConfigDir() string {
	_, filename, _, _ := runtime.Caller(0)
	// 从 internal/data/ 回到 configs/
	return filepath.Join(filepath.Dir(filename), "..", "..", "configs")
}
