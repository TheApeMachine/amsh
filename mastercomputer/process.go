package mastercomputer

import (
	"context"

	"github.com/google/uuid"
)

type Process struct {
	ID     string `json:"id"`
	pctx   context.Context
	ctx    context.Context
	cancel context.CancelFunc
}

func NewProcess(pctx context.Context) *Process {
	return &Process{
		ID:   uuid.New().String(),
		pctx: pctx,
	}
}

func (process *Process) Initialize() {
	process.ctx, process.cancel = context.WithCancel(context.Background())
}

func (process *Process) Run() <-chan string {
	out := make(chan string)

	go func() {
		for {
			select {
			case <-process.pctx.Done():
				process.cancel()
				return
			default:
			}
		}
	}()

	return out
}
