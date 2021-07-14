package eventengine

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Laisky/go-eventengine/internal/consts"
	"github.com/Laisky/go-eventengine/mq/redis"
	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/stretchr/testify/require"
)

func ExampleEventEngine() {
	ctx := context.Background()
	evtstore, err := NewEventEngine(ctx,
		WithEventEngineChanBuffer(1),
		WithEventEngineNFork(2),
		WithEventEngineSuppressPanic(false),
	)
	if err == nil {
		gutils.Logger.Panic("new evt engine", zap.Error(err))
	}

	var (
		topic1 consts.EventTopic = "t1"
		topic2 consts.EventTopic = "t2"
	)
	evt1 := &consts.Event{
		Topic: topic1,
		Meta: consts.EventMeta{
			"name": "yo",
		},
	}
	evt2 := &consts.Event{
		Topic: topic2,
		Meta: consts.EventMeta{
			"name": "yo2",
		},
	}

	handler := func(evt *consts.Event) error {
		fmt.Printf("got event %s: %v\n", evt.Topic, evt.Meta)
		return nil
	}

	evtstore.Register(topic1, handler)
	evtstore.Publish(ctx, evt1) // Output: got event t1: map[name]yo
	evtstore.Publish(ctx, evt2) // nothing print

	evtstore.UnRegister(topic1, handler)
	evtstore.Publish(ctx, evt1) // nothing print
	evtstore.Publish(ctx, evt2) // nothing print

}

func TestNewEventEngine(t *testing.T) {
	ctx := context.Background()
	evtstore, err := NewEventEngine(ctx)
	require.NoError(t, err)

	var (
		topic1 consts.EventTopic = "t1"
		topic2 consts.EventTopic = "t2"
	)
	newEvt1 := func() *consts.Event {
		return &consts.Event{
			Topic: topic1,
			Meta: consts.EventMeta{
				"name": "yo",
			},
		}
	}
	newEvt2 := func() *consts.Event {
		return &consts.Event{
			Topic: topic2,
			Meta: consts.EventMeta{
				"name": "yo2",
			},
		}
	}

	var count int32
	handler := func(evt *consts.Event) error {
		t.Logf("got event %s: %+v", evt.Topic, evt.Meta)
		atomic.AddInt32(&count, 1)
		return nil
	}

	evtstore.Register(topic1, handler)
	evtstore.Publish(ctx, newEvt1())
	evtstore.Publish(ctx, newEvt2())
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, int(atomic.LoadInt32(&count)))

	evtstore.UnRegister(topic1, handler)
	evtstore.Publish(ctx, newEvt1())
	evtstore.Publish(ctx, newEvt2())
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, int(atomic.LoadInt32(&count)))

	evtstore.Register(topic1, handler)
	evtstore.Register(topic2, handler)
	evtstore.Publish(ctx, newEvt1())
	evtstore.Publish(ctx, newEvt2())
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 3, int(atomic.LoadInt32(&count)))
}

func TestNewEventEngineWithMQ(t *testing.T) {
	t.Skipf("need redis")

	ctx := context.Background()

	rdbMQ, err := redis.New()
	require.NoError(t, err)

	evtstore, err := NewEventEngine(ctx,
		WithMQ(rdbMQ))
	require.NoError(t, err)

	var (
		topic1 consts.EventTopic = "t1"
		topic2 consts.EventTopic = "t2"
	)
	newEvt1 := func() *consts.Event {
		return &consts.Event{
			Topic: topic1,
			Meta: consts.EventMeta{
				"name": "yo",
			},
		}
	}
	newEvt2 := func() *consts.Event {
		return &consts.Event{
			Topic: topic2,
			Meta: consts.EventMeta{
				"name": "yo2",
			},
		}
	}

	var target, count int32
	closeCh := make(chan struct{})
	handler := func(evt *consts.Event) error {
		gutils.Logger.Info("handle event",
			zap.String("topic", evt.Topic.String()),
			zap.Any("meta", evt.Meta),
		)
		fmt.Println(atomic.LoadInt32(&target))
		if atomic.AddInt32(&count, 1) == atomic.LoadInt32(&target) {
			closeCh <- struct{}{}
		}

		return nil
	}

	t.Log("case: listen to 1 topic")
	{
		atomic.StoreInt32(&target, 2)
		evtstore.Register(topic1, handler)
		evtstore.Register(topic1, handler)
		evtstore.Register(topic1, handler)
		evtstore.Publish(ctx, newEvt1())
		evtstore.Publish(ctx, newEvt2())
		evtstore.Publish(ctx, newEvt1())
		<-closeCh
		time.Sleep(time.Second)
		require.Equal(t, 2, int(atomic.LoadInt32(&count)))
		evtstore.UnRegister(topic1, handler)
		atomic.StoreInt32(&count, 0)
	}

	t.Log("case: listen to 2 topic")
	{
		atomic.StoreInt32(&target, 4)
		evtstore.Register(topic1, handler)
		evtstore.Register(topic2, handler)
		evtstore.Publish(ctx, newEvt1())
		evtstore.Publish(ctx, newEvt2())
		evtstore.Publish(ctx, newEvt1())
		<-closeCh
		time.Sleep(time.Second)
		require.Equal(t, 4, int(atomic.LoadInt32(&count)))
		atomic.StoreInt32(&count, 0)
	}
}

func BenchmarkNewEventEngine(b *testing.B) {
	ctx := context.Background()
	evtstore, err := NewEventEngine(ctx)
	if err != nil {
		b.Fatalf("%+v", err)
	}

	var (
		topic1 consts.EventTopic = "t1"
		topic2 consts.EventTopic = "t2"
	)
	evt1 := &consts.Event{
		Topic: topic1,
		Meta: consts.EventMeta{
			"name": "yo",
		},
	}
	evt2 := &consts.Event{
		Topic: topic2,
		Meta: consts.EventMeta{
			"name": "yo2",
		},
	}

	handler := func(evt *consts.Event) error {
		b.Logf("got event %s: %+v", evt.Topic, evt.Meta)
		return nil
	}

	evtstore.Register(topic1, handler)

	b.Run("publish", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			evtstore.Publish(ctx, evt1)
			evtstore.Publish(ctx, evt2)
		}
	})

	// b.Error()
}
