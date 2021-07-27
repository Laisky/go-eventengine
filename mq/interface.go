package mq

import (
	"context"

	"github.com/Laisky/go-eventengine/mq/redis"
	"github.com/Laisky/go-eventengine/types"
)

var (
	_ Interface = new(redis.Type)
)

// Interface is MQ backend for event engine
type Interface interface {
	Name() string
	// Put put event into mq
	Put(ctx context.Context, evt *types.Event) error
	// Get get event from mq
	//
	// you can get event from evtChan,
	// if evtChan is closed, then you can get err from errChan.
	Get(ctx context.Context, topic types.EventTopic) (evtChan <-chan *types.Event, errChan <-chan error)
	// Commit mark event as done
	//
	// not all mq support this feature.
	Commit(ctx context.Context, evt *types.Event) error
}
