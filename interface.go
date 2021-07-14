package eventengine

import (
	"context"

	"github.com/Laisky/go-eventengine/internal/consts"
)

type Interface interface {
	// Register register new handler for topic
	//
	// Args:
	//   * topic: specific the topic that will trigger the handler
	//   * handler: the func that used to process event
	//
	// handler will be invoked when eventengine received an event in specific topic.
	//
	// you can register same handler with different topic.
	// if you register same handler with same topic,
	// only one handler will be accepted.
	Register(topic consts.EventTopic, handler EventHandler)
	// UnRegister remove handler for topic
	//
	// if one handler register in different topic,
	// only the handler in specific topic will be removed.
	UnRegister(topic consts.EventTopic, handler EventHandler)
	// Publish publish new event
	Publish(ctx context.Context, evt *consts.Event)
}
