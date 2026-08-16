package biz

import (
	kerrors "github.com/go-kratos/kratos/v2/errors"
)

// errText 将 gRPC 业务错误转换为用户可读的中文提示（供 agent 工具返回给用户）：
//   - kratos error：提取业务服务定义的友好 Message（如「商品库存不足」），
//     避免展示 gRPC 客户端原始错误（形如 "rpc error: code = ... desc = ..."）
//   - 其余（网络/超时/未知）：返回通用文案，不暴露内部细节
func errText(err error) string {
	if err == nil {
		return ""
	}
	if e := kerrors.FromError(err); e != nil {
		if e.Message != "" {
			return e.Message
		}
		return e.Reason
	}
	return "服务暂时不可用，请稍后重试"
}
