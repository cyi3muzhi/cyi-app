package biz

import (
	"context"
	"fmt"
	"sort"
	"strings"

	userRpc "cyi-app/api/service/user/v1"

	"github.com/go-kratos/blades/tools"
)

// RegisterPermissionTool 注册权限查询工具
func RegisterPermissionTool(r *ToolRegistry, roleClient userRpc.AdminRoleClient) {
	t, _ := tools.NewFunc("query_user_permissions", "查询当前登录用户的角色和完整权限列表。当用户问'我有什么权限'、'我能做什么'时调用此工具",
		func(ctx context.Context, _ struct{}) (string, error) {
			role, _ := ctx.Value(roleContextKey).(string)
			if role == "" {
				return "", fmt.Errorf("无法获取当前用户角色信息")
			}

			if roleClient == nil {
				return "", fmt.Errorf("权限服务不可用")
			}

			// 从 user 服务查询角色的模块级权限
			resp, err := roleClient.GetRolePermissions(ctx, &userRpc.GetRolePermissionsRequest{Role: role})
			if err != nil {
				return "", fmt.Errorf("查询权限失败: %v", err)
			}

			if resp == nil || len(resp.Permissions) == 0 {
				return fmt.Sprintf("你的角色：%s\n\n系统中暂未配置具体的模块权限。", role), nil
			}

			// 格式化权限列表
			var modules []string
			for mod := range resp.Permissions {
				modules = append(modules, mod)
			}
			sort.Strings(modules)

			var result string
			result += fmt.Sprintf("你的角色：%s\n\n你的权限：\n", role)
			for _, mod := range modules {
				pv := resp.Permissions[mod]
				if pv == nil || len(pv.Values) == 0 {
					continue
				}
				var actions []string
				for action, allowed := range pv.Values {
					if allowed {
						actions = append(actions, action)
					}
				}
				sort.Strings(actions)
				result += fmt.Sprintf("- %s：%s\n", mod, strings.Join(actions, "、"))
			}
			return result, nil
		})

	t2, _ := tools.NewFunc("check_user_permission", "检查当前用户是否有某个具体操作的权限。参数 module 为模块名（如 product、order），action 为操作名（如 create、view），返回是否有权限",
		func(ctx context.Context, input struct {
			Module string `json:"module"`
			Action string `json:"action"`
		}) (string, error) {
			role, _ := ctx.Value(roleContextKey).(string)
			if role == "" {
				return "", fmt.Errorf("无法获取当前用户角色信息")
			}

			if roleClient == nil {
				return "", fmt.Errorf("权限服务不可用")
			}

			resp, err := roleClient.GetRolePermissions(ctx, &userRpc.GetRolePermissionsRequest{Role: role})
			if err != nil {
				return "", fmt.Errorf("查询权限失败: %v", err)
			}

			if resp != nil && resp.Permissions != nil {
				if pv, ok := resp.Permissions[input.Module]; ok && pv != nil {
					if allowed, ok := pv.Values[input.Action]; ok && allowed {
						return fmt.Sprintf("你的角色「%s」拥有「%s」模块的「%s」权限。", role, input.Module, input.Action), nil
					}
				}
			}

			return fmt.Sprintf("你的角色「%s」没有「%s」模块的「%s」权限。", role, input.Module, input.Action), nil
		})

	r.Add(t)
	r.Add(t2)
}
