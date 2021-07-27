package eventengine

import (
	"context"

	"github.com/Laisky/go-eventengine/types"
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
	Register(topic types.EventTopic, handler Handler)
	// UnRegister remove handler for topic
	//
	// if one handler register in different topic,
	// only the handler in specific topic will be removed.
	UnRegister(topic types.EventTopic, handler Handler)
	// Publish publish new event
	Publish(ctx context.Context, evt *types.Event)
}
