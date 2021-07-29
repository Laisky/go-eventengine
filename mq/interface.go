package mq

import (
	"context"

	"github.com/Laisky/go-eventengine/internal/consts"
)

// Interface is MQ backend for event engine
type Interface interface {
	Name() string
	// Put put event into mq
	Put(ctx context.Context, evt *consts.Event) error
	// Get get event from mq
	//
	// you can get event from evtChan,
	// if evtChan is closed, then you can get err from errChan.
	Get(ctx context.Context, topic consts.EventTopic) (evtChan <-chan *consts.Event, errChan <-chan error)
	// Commit mark event as done
	//
	// not all mq support this feature.
	Commit(ctx context.Context, evt *consts.Event) error
}
