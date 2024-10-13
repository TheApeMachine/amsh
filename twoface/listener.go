package twoface

import (
	"context"

	"github.com/theapemachine/amsh/data"
)

type Listener struct {
	ctx       context.Context
	generator chan data.Artifact
}

func NewListener(ctx context.Context, generator chan data.Artifact) *Listener {
	return &Listener{ctx: ctx, generator: generator}
}

func (listener *Listener) Messages(handlerFunc func(data.Artifact)) {
	go func(fn func(data.Artifact)) {
		for {
			select {
			case msg, ok := <-listener.generator:
				if !ok {
					return
				}
				fn(msg)
			case <-listener.ctx.Done():
				return
			}
		}
	}(handlerFunc)
}
