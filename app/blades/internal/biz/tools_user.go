package biz

import (
	"context"
	"fmt"

	userRpc "cyi-app/api/service/user/v1"
	"github.com/go-kratos/blades/tools"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RegisterUserTools 注册用户工具
func RegisterUserTools(r *ToolRegistry, client userRpc.UserClient) {
	// query_users 用户总数统计
	t1, _ := tools.NewFunc("query_users", "查询平台用户总数和统计信息",
		func(ctx context.Context, input struct{}) (string, error) {
			reply, err := client.UserCount(ctx, &emptypb.Empty{})
			if err != nil {
				return errText(err), nil
			}
			return fmt.Sprintf("平台总用户数：%d", reply.Count), nil
		})
	r.Add(t1)
}
