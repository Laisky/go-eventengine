// Package mock is mock mq for unit test
package mock

import (
	"context"
	"sync"

	"github.com/Laisky/go-eventengine/types"
)

type Type struct {
	sync.Mutex
	topic2evtq map[types.EventTopic]chan *types.Event
}

func New() (*Type, error) {
	return &Type{
		topic2evtq: map[types.EventTopic]chan *types.Event{},
	}, nil
}

func (t *Type) Name() string {
	return "mock"
}

func (t *Type) Put(ctx context.Context, evt *types.Event) error {
	t.Lock()
	ch, ok := t.topic2evtq[evt.Topic]
	if !ok {
		ch = make(chan *types.Event, 1000)
		t.topic2evtq[evt.Topic] = ch
	}
	t.Unlock()

	ch <- evt
	return nil
}

func (t *Type) Get(ctx context.Context, topic types.EventTopic) (evtChan <-chan *types.Event, errChan <-chan error) {
	t.Lock()
	ch, ok := t.topic2evtq[topic]
	if !ok {
		ch = make(chan *types.Event, 1000)
		t.topic2evtq[topic] = ch
	}
	t.Unlock()

	return ch, make(<-chan error)
}

func (t *Type) Commit(ctx context.Context, evt *types.Event) error {
	return nil
}
