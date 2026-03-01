package asynqhelper

import (
	"github.com/redis/go-redis/v9"
)

type queueRedis struct {
	redis.UniversalClient
}

func NewQueueRedis(c redis.UniversalClient) *queueRedis {
	return &queueRedis{
		UniversalClient: c,
	}
}

func (q *queueRedis) MakeRedisClient() interface{} {
	return q
}
