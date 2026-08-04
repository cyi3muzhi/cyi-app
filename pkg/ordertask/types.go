// Package ordertask 提供订单延迟任务的共享类型与常量。
// 订单服务（投递端）与 task 服务（消费端）共同引用本包，
// 避免跨服务反向依赖 biz 包，并保证两端任务类型名与 payload 结构一致。
package ordertask

import (
	"encoding/json"
	"fmt"
	"time"
)

// OrderTimeoutCancelName 订单超时取消任务的逻辑名称（不含 delay: 前缀）
// task 服务通过 DelayHandleFunc 注册时使用本常量，transporter.DelayName 会自动补 "delay:" 前缀。
// 注意：不要在此带 "delay:" 前缀，否则会变成 "delay:delay:order_timeout_cancel" 导致 handler not found。
const OrderTimeoutCancelName = "order_timeout_cancel"

// OrderTimeoutTaskType 订单超时取消任务在 asynq 队列中的完整类型名
// 订单服务投递任务时使用（asynq.NewTask 的第一个参数）。
// 格式 "delay:{name}" 与 transporter.DelayName(OrderTimeoutCancelName) 的产物保持一致，
// 保证 task 注册的 handler name 与订单服务投递的 task type 完全对齐。
const OrderTimeoutTaskType = "delay:" + OrderTimeoutCancelName

// OrderTimeoutDelay 订单超时时间：待付款订单超过该时长未支付则自动取消
const OrderTimeoutDelay = 30 * time.Minute

// OrderTimeoutPayload Asynq 延迟任务 payload（序列化为 JSON 传递给 worker）
// 订单服务投递时构造，task 消费时反序列化。
type OrderTimeoutPayload struct {
	OrderID int64 `json:"order_id"`
}

// TaskID 构造订单超时任务的唯一 TaskID（同一订单仅允许存在一个待执行的超时任务）
func TaskID(orderID int64) string {
	return fmt.Sprintf("order_timeout:%d", orderID)
}

// MarshalPayload 序列化任务 payload（供订单服务投递 asynq.NewTask 时使用）
func MarshalPayload(p *OrderTimeoutPayload) ([]byte, error) {
	return json.Marshal(p)
}

// UnmarshalPayload 反序列化任务 payload（供 task 消费 asynq.Task.Payload() 时使用）
func UnmarshalPayload(b []byte) (*OrderTimeoutPayload, error) {
	var p OrderTimeoutPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("解析订单超时payload失败: %w", err)
	}
	return &p, nil
}
