package server

import (
	v1 "cyi-app/api/job/v1"
	"cyi-app/app/job/internal/service"
	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/hibiken/asynq"
)

func ProcessTask(m *ts.TaskServer, job *service.JobService) *asynq.ServeMux {
	m.CronHandleFunc(v1.JobName_user_count.String(), job.UserCount)
	m.CronHandleFunc("coupon_expire", job.CouponExpire)
	return m.TaskServer
}
