package biz

import "context"

// WithUserID 将用户 ID 写入 context
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// GetUserID 从 context 读取用户 ID
func GetUserID(ctx context.Context) int64 {
	if v, ok := ctx.Value(userIDContextKey).(int64); ok {
		return v
	}
	return 0
}

// WithSource 将来源写入 context
func WithSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, sourceContextKey, source)
}

// GetSource 从 context 读取来源
func GetSource(ctx context.Context) string {
	if v, ok := ctx.Value(sourceContextKey).(string); ok {
		return v
	}
	return ""
}

// WithRole 将角色写入 context
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleContextKey, role)
}

// GetRole 从 context 读取角色
func GetRole(ctx context.Context) string {
	if v, ok := ctx.Value(roleContextKey).(string); ok {
		return v
	}
	return ""
}
