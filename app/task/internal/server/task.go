package server

import (
	v1 "cyi-app/api/task/v1"
	"cyi-app/app/task/internal/service"
	"cyi-app/pkg/exporttask"
	"cyi-app/pkg/ordertask"
	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/hibiken/asynq"
)

// ProcessTask 构建 asynq 路由 mux，注册各任务处理器
func ProcessTask(m *ts.TaskServer, task *service.TaskService) *asynq.ServeMux {
	m.AsyncHandleFunc("email:send", task.EmailTask)
	m.AsyncHandleFunc(v1.TaskName_create_user.String(), task.CreateUserTask)
	// 订单导出任务（由 admin 投递到 export 队列，task 消费）
	m.AsyncHandleFunc(exporttask.ExportOrderTaskName, task.HandleExportOrderTask)
	// 订单超时自动取消任务（由订单服务投递延迟任务，30 分钟后消费）
	m.DelayHandleFunc(ordertask.OrderTimeoutCancelName, task.OrderTimeoutCancel)
	// cron 定时任务（由 Scheduler 投递到 critical 队列，task 消费）
	m.CronHandleFunc(v1.TaskName_user_count.String(), task.UserCount)
	m.CronHandleFunc(v1.TaskName_coupon_expire.String(), task.CouponExpire)
	// 订单超时兜底扫描（cron 定时任务，周期性取消超时未支付订单）
	m.CronHandleFunc(orderTimeoutCheckName, task.OrderTimeoutScan)
	return m.TaskServer
}

// orderTimeoutCheckName 订单超时兜底扫描任务的 cron 名称（与 config.yaml tasks 段 name 对应）
const orderTimeoutCheckName = "order_timeout_check"
