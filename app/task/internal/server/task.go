package server

import (
	v1 "cyi-app/api/task/v1"
	"cyi-app/app/task/internal/service"
	"cyi-app/pkg/exporttask"
	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/hibiken/asynq"
)

// ProcessTask 构建 asynq 路由 mux，注册各任务处理器
func ProcessTask(m *ts.TaskServer, task *service.TaskService) *asynq.ServeMux {
	m.AsyncHandleFunc("email:send", task.EmailTask)
	m.AsyncHandleFunc(v1.TaskName_create_user.String(), task.CreateUserTask)
	// 订单导出任务（由 admin 投递到 export 队列，task 消费）
	m.AsyncHandleFunc(exporttask.ExportOrderTaskName, task.HandleExportOrderTask)
	return m.TaskServer
}
