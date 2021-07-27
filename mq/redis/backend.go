package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/Laisky/go-eventengine/internal/consts"
	"github.com/Laisky/go-eventengine/types"
	gredis "github.com/Laisky/go-redis"
	gutils "github.com/Laisky/go-utils"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// Type mq base on redis
type Type struct {
	rdbKeyPrefix string
	rdb          *gredis.Utils
	logger       *gutils.LoggerType
}

// OptFunc optional arguments
type OptFunc func(t *Type) error

// WithRDBCli set redis client
func WithRDBCli(rdb *redis.Client) OptFunc {
	return func(t *Type) error {
		if rdb == nil {
			return errors.Errorf("rdb is nil")
		}

		t.rdb = gredis.NewRedisUtils(rdb)
		return nil
	}
}

// WithLogger set internal logger
func WithLogger(l *gutils.LoggerType) OptFunc {
	return func(t *Type) error {
		if l == nil {
			return errors.Errorf("logger is nil")
		}

		t.logger = l
		return nil
	}
}

// WithKeyPrefix set redis key prefix
func WithKeyPrefix(prefix string) OptFunc {
	return func(t *Type) error {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			return errors.Errorf("prefix is empty")
		}

		if prefix[len(prefix)-1] != '/' {
			prefix += "/"
		}

		t.rdbKeyPrefix = prefix
		return nil
	}
}

// New create mq depends on redis
func New(optfs ...OptFunc) (t *Type, err error) {
	t = &Type{
		rdbKeyPrefix: "/go-eventengine/",
	}
	for _, optf := range optfs {
		if err := optf(t); err != nil {
			return nil, err
		}
	}

	if t.rdb == nil {
		t.rdb = gredis.NewRedisUtils(redis.NewClient(new(redis.Options)))
	}

	if t.logger == nil {
		if t.logger, err = gutils.NewConsoleLoggerWithName("regis", gutils.LoggerLevelInfo); err != nil {
			return nil, err
		}
	}

	return t, nil
}

// Name get name of mq
func (t *Type) Name() string {
	return "redis"
}

// RDBKey generate redis key by topic
func (t *Type) RDBKey(topic types.EventTopic) string {
	return t.rdbKeyPrefix + consts.RedisKeyQueue + topic.String() + "/"
}

// Put put event into mq
func (t *Type) Put(ctx context.Context, evt *types.Event) error {
	msg, err := gutils.JSON.MarshalToString(evt)
	if err != nil {
		return err
	}

	return t.rdb.RPush(ctx, t.RDBKey(evt.Topic), msg)
}

// Get get evt from mq
func (t *Type) Get(ctx context.Context, topic types.EventTopic) (<-chan *types.Event, <-chan error) {
	eventChan := make(chan *types.Event)
	errChan := make(chan error)
	go func() {
		defer close(eventChan)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			_, v, err := t.rdb.LPopKeysBlocking(ctx, t.RDBKey(topic))
			if err != nil {
				errChan <- err
				return
			}

			evt := new(types.Event)
			if err = gutils.JSON.UnmarshalFromString(v, evt); err != nil {
				errChan <- err
				return
			}

			fmt.Println("get evt", v)
			eventChan <- evt
		}
	}()

	return eventChan, errChan
}

// FIXME: not implement
func (t *Type) Commit(ctx context.Context, evt *types.Event) error {
	return errors.New("NotImplement")
}
