# go-eventengine

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Commitizen friendly](https://img.shields.io/badge/commitizen-friendly-brightgreen.svg)](http://commitizen.github.io/cz-cli/)
[![Go Report Card](https://goreportcard.com/badge/github.com/Laisky/go-eventengine)](https://goreportcard.com/report/github.com/Laisky/go-eventengine)
[![GoDoc](https://godoc.org/github.com/Laisky/go-eventengine?status.svg)](https://pkg.go.dev/github.com/Laisky/go-eventengine?tab=doc)
[![Build Status](https://travis-ci.com/Laisky/go-eventengine.svg?branch=main)](https://travis-ci.com/Laisky/go-eventengine)
[![codecov](https://codecov.io/gh/Laisky/go-eventengine/branch/main/graph/badge.svg)](https://codecov.io/gh/Laisky/go-eventengine)


simple event driven tools


## Example

```go
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

```
