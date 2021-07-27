package main

import (
	"context"

	eventengine "github.com/Laisky/go-eventengine"
	"github.com/Laisky/go-eventengine/types"
	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
)

var closed = make(chan struct{})

func handler(evt *types.Event) error {
	gutils.Logger.Info("handler", zap.Any("event", evt))
	closed <- struct{}{}
	return nil
}

func main() {
	ctx := context.Background()

	engine, err := eventengine.New(ctx)
	if err != nil {
		gutils.Logger.Panic("new engine", zap.Error(err))
	}

	var topic types.EventTopic = "hello"

	// register handler to topic
	engine.Register(topic, handler)

	// trigger event
	engine.Publish(ctx, &types.Event{Topic: topic})

	// wait for handler to be called
	<-closed
}
