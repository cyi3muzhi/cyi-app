package server

import (
	v1 "cyi-app/api/task/v1"
	"cyi-app/app/task/internal/service"
	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/hibiken/asynq"
)

func ProcessTask(m *ts.TaskServer, task *service.TaskService) *asynq.ServeMux {
	m.AsyncHandleFunc("email:send", task.EmailTask)
	m.AsyncHandleFunc(v1.TaskName_create_user.String(), task.CreateUserTask)
	return m.TaskServer
}
