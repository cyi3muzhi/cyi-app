package server

import (
	"cyi-app/app/task/internal/service"
	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/hibiken/asynq"
)

func ProcessTask(m *ts.TaskServer, task *service.TaskService) *asynq.ServeMux {
	m.TaskServer.HandleFunc(ts.AsyncName("email:send"), task.EmailTask)
	return m.TaskServer
}
