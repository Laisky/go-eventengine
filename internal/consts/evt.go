package consts

import (
	"errors"
	"time"
)

// EventTopic topic of event
type EventTopic string

func (h EventTopic) String() string {
	return string(h)
}

// HandlerID id(name) of event handler
//
//   first character of HandlerID should not be `@`.
type HandlerID string

func (h HandlerID) String() string {
	return string(h)
}

// Valid parse string to handler id with validation
func (h HandlerID) Valid() error {
	if len(h) == 0 {
		return errors.New("handler id is empty")
	}

	if h[0] == '@' {
		return errors.New("first character of handler id should not be `@`")
	}

	return nil
}

// MetaKey key of event's meta
type MetaKey string

func (h MetaKey) String() string {
	return string(h)
}

// EventMeta event meta
type EventMeta map[MetaKey]interface{}

// Event evt
type Event struct {
	Topic EventTopic
	Time  time.Time
	Meta  EventMeta
	Stack string
}
