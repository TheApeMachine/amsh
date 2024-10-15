package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
)

type Messaging struct {
	pctx     context.Context
	ctx      context.Context
	cancel   context.CancelFunc
	queue    *twoface.Queue
	inbox    chan data.Artifact
	messages []data.Artifact
}

func NewMessaging() *Messaging {
	return &Messaging{
		queue: twoface.NewQueue(),
		inbox: make(chan data.Artifact, 128),
	}
}

func (messaging *Messaging) Initialize(ctx context.Context) error {
	messaging.pctx = ctx
	messaging.ctx, messaging.cancel = context.WithCancel(messaging.pctx)

	go func() {
		for {
			select {
			case <-messaging.pctx.Done():
				return
			case <-messaging.ctx.Done():
				return
			case artifact := <-messaging.inbox:
				messaging.messages = append(messaging.messages, artifact)
			}
		}
	}()

	return nil
}

/*
Read implements the io.Reader interface for the messaging.
*/
func (messaging *Messaging) Read(p []byte) (n int, err error) {
	return 0, nil
}

/*
Write implements the io.Writer interface for the messaging.
*/
func (messaging *Messaging) Write(p []byte) (n int, err error) {
	if artifact := data.Empty.Unmarshal(p); artifact != data.Empty {
		messaging.queue.Write(p)

	}
	return 0, nil
}

/*
Close implements the io.Closer interface for the messaging.
*/
func (messaging *Messaging) Close() error {
	messaging.cancel()
	return nil
}
