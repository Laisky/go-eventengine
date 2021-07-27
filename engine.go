package eventengine

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/Laisky/go-eventengine/mq"
	"github.com/Laisky/go-eventengine/types"
	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/Laisky/zap/zapcore"
	"github.com/pkg/errors"
)

const (
	defaultNFork         int = 2
	defaultMsgBufferSize int = 0
)

// Handler function to handle event
type Handler func(*types.Event) error

// Type event driven engine
//
// Usage
//
// you -> produce event -> trigger multiply handlers
//
//   1. create an engine by `NewEventEngine`
//   2. register handlers with specified event type by `engine.Register`
//   3. produce event to trigger handlers by `engine.Publish`
type Type struct {
	*eventStoreManagerOpt

	// topic2hs map[topic]*sync.Map[handlerID]handler
	topic2hs *sync.Map
	taskChan chan *eventRunChanItem

	// -------------------------------------
	// mq
	// -------------------------------------

	mqAddTopic     chan types.EventTopic
	mqRemoveTopic  chan types.EventTopic
	mqTopic2Cancel map[types.EventTopic]context.CancelFunc
}

type eventStoreManagerOpt struct {
	msgBufferSize int
	nfork         int
	logger        *gutils.LoggerType
	suppressPanic bool
	mq            mq.Interface
}

// OptFunc options for EventEngine
type OptFunc func(*eventStoreManagerOpt) error

// WithNFork set nfork of event store
//
// default to 2
func WithNFork(nfork int) OptFunc {
	return func(opt *eventStoreManagerOpt) error {
		if nfork <= 0 {
			return errors.Errorf("nfork must > 0")
		}

		opt.nfork = nfork
		return nil
	}
}

// WithChanBuffer set msg buffer size of event store
//
// default to 1
func WithChanBuffer(msgBufferSize int) OptFunc {
	return func(opt *eventStoreManagerOpt) error {
		if msgBufferSize < 0 {
			return errors.Errorf("msgBufferSize must >= 0")
		}

		opt.msgBufferSize = msgBufferSize
		return nil
	}
}

// WithLogger set event store's logger
//
// default to gutils' internal logger
func WithLogger(logger *gutils.LoggerType) OptFunc {
	return func(opt *eventStoreManagerOpt) error {
		if logger == nil {
			return errors.Errorf("logger is nil")
		}

		opt.logger = logger
		return nil
	}
}

// WithSuppressPanic set whether suppress event handler's panic
//
// default to false
func WithSuppressPanic(suppressPanic bool) OptFunc {
	return func(opt *eventStoreManagerOpt) error {
		opt.suppressPanic = suppressPanic
		return nil
	}
}

// WithMQ set whether suppress event handler's panic
//
// default to null
func WithMQ(mq mq.Interface) OptFunc {
	return func(opt *eventStoreManagerOpt) error {
		if mq == nil {
			return errors.Errorf("mq is nil")
		}

		opt.mq = mq
		return nil
	}
}

// New new event store manager
//
// Args:
//   * ctx:
//   * WithNFork: n goroutines to run handlers in parallel
//   * WithChanBuffer: length of channel to receive published event
//   * WithLogger: internal logger in event engine
//   * WithSuppressPanic: if is true, will not raise panic when running handler
func New(ctx context.Context, opts ...OptFunc) (Interface, error) {
	opt := &eventStoreManagerOpt{
		msgBufferSize: defaultMsgBufferSize,
		nfork:         defaultNFork,
		logger:        gutils.Logger.Named("evt-store-" + gutils.RandomStringWithLength(6)),
	}
	for _, optf := range opts {
		if err := optf(opt); err != nil {
			return nil, err
		}
	}

	e := &Type{
		eventStoreManagerOpt: opt,
		topic2hs:             &sync.Map{},
		taskChan:             make(chan *eventRunChanItem, opt.msgBufferSize),

		mqAddTopic:     make(chan types.EventTopic),
		mqRemoveTopic:  make(chan types.EventTopic),
		mqTopic2Cancel: map[types.EventTopic]context.CancelFunc{},
	}

	e.runHandlerRunner(ctx, opt.nfork)
	go e.runMQListener(ctx)

	fields := []zapcore.Field{
		zap.Int("nfork", opt.nfork),
		zap.Int("buffer", opt.msgBufferSize),
	}
	if e.mq != nil {
		fields = append(fields, zap.String("mq", e.mq.Name()))
	}
	e.logger.Info("new event store", fields...)
	return e, nil
}

func runHandlerWithoutPanic(h Handler, evt *types.Event) (err error) {
	defer func() {
		if erri := recover(); erri != nil {
			err = errors.Errorf("run event handler with evt `%s`: %+v", evt.Topic, erri)
		}
	}()

	err = h(evt)
	return err
}

type eventRunChanItem struct {
	h   Handler
	hid types.HandlerID
	evt *types.Event
}

func (e *Type) runHandlerRunner(ctx context.Context, nfork int) {
	for i := 0; i < nfork; i++ {
		logger := e.logger.Named(strconv.Itoa(i))
		go func() {
			var err error
			for {
				select {
				case <-ctx.Done():
					return
				case t := <-e.taskChan:
					logger.Debug("trigger handler",
						zap.String("evt", t.evt.Topic.String()),
						zap.String("source", t.evt.Stack),
						zap.String("handler", t.hid.String()))

					if e.suppressPanic {
						err = runHandlerWithoutPanic(t.h, t.evt)
					} else {
						err = t.h(t.evt)
					}

					if err != nil {
						logger.Error("run evnet handler",
							zap.String("evt", t.evt.Topic.String()),
							zap.String("handler", t.hid.String()),
							zap.String("source", t.evt.Stack),
							zap.Error(err))
					}
				}
			}
		}()
	}
}

const (
	handlerIDput2mq types.HandlerID = "@put2mq"
)

// put2mq put event into mq
func (e *Type) put2mq(evt *types.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.mq.Put(ctx, evt); err != nil {
		return errors.Wrap(err, "publish evt to mq")
	}

	return nil
}

// runMqListener fetch event from MQ
func (e *Type) runMQListener(ctx context.Context) {
	if e.mq == nil {
		return
	}

	for {
		select {
		case topic := <-e.mqAddTopic:
			// check whether there is already has a listener to this topic
			if _, ok := e.mqTopic2Cancel[topic]; ok {
				continue
			}

			// create new listener
			ctx2mq, cancel := context.WithCancel(ctx)
			e.mqTopic2Cancel[topic] = cancel
			go func() {
				defer cancel()
				evtChan, errChan := e.mq.Get(ctx2mq, topic)
			EVT_LOOP:
				for {
					select {
					case evt, ok := <-evtChan:
						if !ok {
							break EVT_LOOP
						}

						e.triggerHandler(evt)
					case <-ctx2mq.Done():
						break EVT_LOOP
					}
				}

				if err := <-errChan; err != nil {
					if !errors.Is(err, context.Canceled) {
						e.logger.Error("mq closed", zap.Error(err), zap.String("topic", topic.String()))
					}
				}
			}()
			e.logger.Info("add mq listener",
				zap.String("mq", e.mq.Name()),
				zap.String("topic", topic.String()))
		case topic := <-e.mqRemoveTopic:
			// check whether there is no handler listen to this topic
			if hsi, _ := e.topic2hs.Load(topic); hsi != nil {
				empty := true
				hsi.(*sync.Map).Range(func(key, value interface{}) bool {
					empty = false
					return false
				})

				if !empty {
					continue
				}
			}

			// remove listener
			e.mqTopic2Cancel[topic]()
			delete(e.mqTopic2Cancel, topic)
			e.logger.Info("remove mq listener", zap.String("topic", topic.String()))
		}
	}
}

// Register register handler
func (e *Type) Register(topic types.EventTopic, handler Handler) {
	handlerID := GetHandlerID(handler)
	hs := &sync.Map{}
	actual, _ := e.topic2hs.LoadOrStore(topic, hs)
	actual.(*sync.Map).Store(handlerID, handler)

	// test
	{
		hsi, ok := e.topic2hs.Load(topic)
		if !ok {
			e.logger.Panic("not ok")
		}

		hi, ok := hsi.(*sync.Map).Load(handlerID)
		if !ok {
			e.logger.Panic("not ok")
		}

		fmt.Println(hi)
	}

	if e.mq != nil {
		e.mqAddTopic <- topic
	}

	e.logger.Info("register handler",
		zap.String("topic", topic.String()),
		zap.String("handler", handlerID.String()))
}

// UnRegister unregister handler
func (e *Type) UnRegister(topic types.EventTopic, handler Handler) {
	handlerID := GetHandlerID(handler)
	if hsi, _ := e.topic2hs.Load(topic); hsi != nil {
		hsi.(*sync.Map).Delete(handlerID)
	}

	if e.mq != nil {
		e.mqRemoveTopic <- topic
	}

	e.logger.Info("unregister handler",
		zap.String("topic", topic.String()),
		zap.String("handler", handlerID.String()))
}

func (e *Type) triggerHandler(evt *types.Event) {
	hsi, ok := e.topic2hs.Load(evt.Topic)
	if !ok || hsi == nil {
		return
	}

	hsi.(*sync.Map).Range(func(hid, h interface{}) bool {
		e.taskChan <- &eventRunChanItem{
			h:   h.(Handler),
			hid: hid.(types.HandlerID),
			evt: evt,
		}

		return true
	})
}

// Publish publish new event
func (e *Type) Publish(ctx context.Context, evt *types.Event) {
	e.logger.Debug("publish event", zap.String("event", evt.Topic.String()))
	evt.Time = gutils.Clock.GetUTCNow()
	evt.Stack = string(debug.Stack())

	if e.mq != nil {
		// put event into mq
		e.taskChan <- &eventRunChanItem{
			h:   e.put2mq,
			hid: handlerIDput2mq,
			evt: evt,
		}
		return
	}

	e.triggerHandler(evt)
}
