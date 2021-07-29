package mq

import (
	"github.com/Laisky/go-eventengine/mq/mock"
	"github.com/Laisky/go-eventengine/mq/redis"
)

func WithRedisMQ(optfs ...redis.OptFunc) (Interface, error) {
	return redis.New(optfs...)
}

func WithMockMQ() (Interface, error) {
	return mock.New()
}
