package sockpuppet

import (
	"context"
	"io"
	"sync"
)

/*
Pipe is a bidirectional stream that allows multiple channels (origins and destinations).
It routes data between multiple io.ReadWriteCloser interfaces and supports context for cancellation.
*/
type Pipe struct {
	mu       sync.Mutex
	channels []io.ReadWriteCloser
	errChan  chan error
	ctx      context.Context
	cancel   context.CancelFunc
}

/*
NewPipe creates a new Pipe with optional io.ReadWriteCloser channels and a context for cancellation or timeout.
*/
func NewPipe(ctx context.Context, channels ...io.ReadWriteCloser) *Pipe {
	pipeCtx, cancel := context.WithCancel(ctx)

	pipe := &Pipe{
		channels: channels,
		errChan:  make(chan error, len(channels)),
		ctx:      pipeCtx,
		cancel:   cancel,
	}

	// Start a goroutine to handle data routing between channels
	go pipe.routeData()

	return pipe
}

/*
routeData starts the routing of data between the channels.
Each channel has its own goroutine for reading and writing data.
The function listens to the context for cancellation or timeout.
*/
func (pipe *Pipe) routeData() {
	var wg sync.WaitGroup

	for _, ch := range pipe.channels {
		wg.Add(1)
		go func(ch io.ReadWriteCloser) {
			defer wg.Done()
			pipe.handleChannel(ch)
		}(ch)
	}

	// Wait for all routines to complete, or for the context to be canceled
	wg.Wait()
	close(pipe.errChan)
}

/*
handleChannel reads from a channel and writes to all other channels, respecting the provided context for cancellation or timeout.
*/
func (pipe *Pipe) handleChannel(src io.ReadWriteCloser) {
	buffer := make([]byte, 1024)

	for {
		select {
		// Handle context cancellation or timeout
		case <-pipe.ctx.Done():
			return
		default:
			// Read from the current channel
			n, err := src.Read(buffer)
			if err != nil {
				if err != io.EOF {
					pipe.errChan <- err
				}
				return
			}

			pipe.mu.Lock()
			// Write to all other channels
			for _, dst := range pipe.channels {
				if dst != src {
					if _, err := dst.Write(buffer[:n]); err != nil {
						pipe.errChan <- err
					}
				}
			}
			pipe.mu.Unlock()
		}
	}
}

/*
Close closes all the channels and cancels the context, stopping the pipe's operation.
*/
func (pipe *Pipe) Close() error {
	pipe.cancel() // Cancel the context to stop operations
	for _, ch := range pipe.channels {
		if err := ch.Close(); err != nil {
			return err
		}
	}
	return nil
}

/*
Errors returns a channel where routing errors can be collected.
*/
func (pipe *Pipe) Errors() <-chan error {
	return pipe.errChan
}
