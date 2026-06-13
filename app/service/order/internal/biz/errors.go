package biz

import (
	"fmt"
)

// OrderErrorCode 订单错误码
type OrderErrorCode int

const (
	// ErrOrderAlreadyPaid 订单已支付
	ErrOrderAlreadyPaid OrderErrorCode = 1001
	// ErrOrderNotPaid 订单未支付
	ErrOrderNotPaid OrderErrorCode = 1002
	// ErrOrderAlreadyShipped 订单已发货
	ErrOrderAlreadyShipped OrderErrorCode = 1003
	// ErrOrderNotShipped 订单未发货
	ErrOrderNotShipped OrderErrorCode = 1004
	// ErrOrderAlreadyCompleted 订单已完成
	ErrOrderAlreadyCompleted OrderErrorCode = 1005
	// ErrOrderAlreadyCancelled 订单已取消
	ErrOrderAlreadyCancelled OrderErrorCode = 1006
	// ErrOrderAlreadyRefunding 订单退款中
	ErrOrderAlreadyRefunding OrderErrorCode = 1007
	// ErrOrderAlreadyRefunded 订单已退款
	ErrOrderAlreadyRefunded OrderErrorCode = 1008
	// ErrOrderInvalidState 订单状态无效
	ErrOrderInvalidState OrderErrorCode = 1009
)

// OrderError 订单错误
type OrderError struct {
	Code    OrderErrorCode
	Message string
	Context map[string]interface{}
}

func (e *OrderError) Error() string {
	if e.Context != nil && len(e.Context) > 0 {
		return fmt.Sprintf("[%d] %s: %+v", e.Code, e.Message, e.Context)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// 错误消息映射（可以支持国际化）
var errorMessages = map[OrderErrorCode]string{
	ErrOrderAlreadyPaid:     "订单已支付，无需重复支付",
	ErrOrderNotPaid:         "订单未支付，无法执行该操作",
	ErrOrderAlreadyShipped:  "订单已发货，无需重复发货",
	ErrOrderNotShipped:      "订单未发货，无法执行该操作",
	ErrOrderAlreadyCompleted: "订单已完成，无需重复操作",
	ErrOrderAlreadyCancelled: "订单已取消，无法执行该操作",
	ErrOrderAlreadyRefunding: "订单已在退款中，无需重复操作",
	ErrOrderAlreadyRefunded:  "订单已退款，无需重复操作",
	ErrOrderInvalidState:     "订单状态无效，无法执行该操作",
}

// NewOrderError 创建订单错误
func NewOrderError(code OrderErrorCode, context ...map[string]interface{}) *OrderError {
	ctx := map[string]interface{}{}
	if len(context) > 0 {
		ctx = context[0]
	}

	return &OrderError{
		Code:    code,
		Message: errorMessages[code],
		Context: ctx,
	}
}
