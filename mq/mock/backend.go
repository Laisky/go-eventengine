// Package mock is mock mq for unit test
package mock

import (
	"context"
	"sync"

	"github.com/Laisky/go-eventengine/internal/consts"
)

type Type struct {
	sync.Mutex
	topic2evtq map[consts.EventTopic]chan *consts.Event
}

func New() (*Type, error) {
	return &Type{
		topic2evtq: map[consts.EventTopic]chan *consts.Event{},
	}, nil
}

func (t *Type) Name() string {
	return "mock"
}

func (t *Type) Put(ctx context.Context, evt *consts.Event) error {
	t.Lock()
	ch, ok := t.topic2evtq[evt.Topic]
	if !ok {
		ch = make(chan *consts.Event, 1000)
		t.topic2evtq[evt.Topic] = ch
	}
	t.Unlock()

	ch <- evt
	return nil
}

func (t *Type) Get(ctx context.Context, topic consts.EventTopic) (evtChan <-chan *consts.Event, errChan <-chan error) {
	t.Lock()
	ch, ok := t.topic2evtq[topic]
	if !ok {
		ch = make(chan *consts.Event, 1000)
		t.topic2evtq[topic] = ch
	}
	t.Unlock()

	return ch, make(<-chan error)
}

func (t *Type) Commit(ctx context.Context, evt *consts.Event) error {
	return nil
}
