package biz

import "errors"

// 业务层错误定义（避免依赖 proto/conf 等底层包）
var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("user not found")
	// ErrPasswordError 密码错误
	ErrPasswordError = errors.New("password error")
	// ErrUserAlreadyRegistered 用户已注册
	ErrUserAlreadyRegistered = errors.New("user already registered")
	// ErrUserRegisterFailed 用户注册失败
	ErrUserRegisterFailed = errors.New("user register failed")
	// ErrAddressNotFound 收货地址不存在
	ErrAddressNotFound = errors.New("address not found")
	// ErrMessageNotFound 消息不存在
	ErrMessageNotFound = errors.New("message not found")
	// ErrMessageCreateFailed 消息创建失败
	ErrMessageCreateFailed = errors.New("message create failed")
)

// LoginReply 登录响应（biz 层定义，避免依赖 proto）
type LoginReply struct {
	Token string
}
