package rmq

// 队列名称常量
const (
	QueueOrderCreated = "order_created_queue"
	QueueOrderTimeout = "order_timeout_queue"
	QueueOrderDelay   = "order_delay_queue"

	QueueLogsAll   = "all_logs"
	QueueLogsError = "error_logs"

	QueueProductPrice = "product_price_updates"

	QueuePriorityOrder = "priority_order_queue"

	QueueNormalOrder = "normal_order_queue"
	QueueDLXOrder    = "dlx_order_queue"
)

// 交换机名称常量
const (
	ExchangeLogs        = "logs_exchange"
	ExchangeProduct     = "product_exchange"
	ExchangeDelayTarget = "delay_target_"
	ExchangeDLX         = "dlx_exchange_"
)

// 路由键常量
const (
	RoutingKeyOrderCreated   = "order.created"
	RoutingKeyOrderPaid      = "order.paid"
	RoutingKeyOrderCancelled = "order.cancelled"

	RoutingKeyAllLogs   = "#"
	RoutingKeyErrorLogs = "*.error"

	RoutingKeyDelay      = "delay_key"
	RoutingKeyDeadLetter = "dead_letter"
)

// 消息类型常量
const (
	ContentTypeJSON = "application/json"
	ContentTypeText = "text/plain"
)

// 配置常量
const (
	DefaultMaxPriority  = 10
	DefaultOrderTimeout = 30 // 分钟
)
