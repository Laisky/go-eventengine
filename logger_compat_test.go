package eventengine

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Laisky/go-eventengine/types"
	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/Laisky/zap/zapcore"
	"github.com/stretchr/testify/require"
)

// TestLoggerInterfaceEventDelivery verifies current logger construction supports public event delivery.
func TestLoggerInterfaceEventDelivery(t *testing.T) {
	var logs int32
	logger, err := gutils.NewConsoleLoggerWithName("consumer-contract", gutils.LoggerLevelDebug, zap.Hooks(func(zapcore.Entry) error {
		atomic.AddInt32(&logs, 1)
		return nil
	}))
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine, err := New(ctx, WithLogger(logger))
	require.NoError(t, err)
	delivered := make(chan *types.Event, 1)
	topic := types.EventTopic("logger-contract")
	engine.Register(topic, func(event *types.Event) error {
		delivered <- event
		return nil
	})
	expected := &types.Event{Topic: topic, Meta: types.EventMeta{"marker": "synthetic"}}
	engine.Publish(ctx, expected)
	select {
	case got := <-delivered:
		require.Equal(t, expected, got)
	case <-time.After(5 * time.Second):
		t.Fatal("public event was not delivered")
	}
	require.Greater(t, atomic.LoadInt32(&logs), int32(0), "injected logger must receive engine logs")
}

// TestLoggerInterfaceRejectsNil preserves nil and legacy typed-nil option behavior.
func TestLoggerInterfaceRejectsNil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := New(ctx, WithLogger(nil))
	require.Error(t, err)
	var legacy *gutils.LoggerType
	_, err = New(ctx, WithLogger(legacy))
	require.Error(t, err)
}
