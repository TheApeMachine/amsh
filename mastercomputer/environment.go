package mastercomputer

import (
	"context"
	"io"
	"sync"
)

/*
Environment represents a fully functional Linux environment that can be used to run commands.
It provides a containerized approach to ensure a consistent and isolated environment.
The Environment struct implements io.Reader, io.Writer, and io.Closer interfaces
for seamless interaction with the containerized environment.
*/
type Environment struct {
	ctx    context.Context
	cancel context.CancelFunc
	stdin  io.WriteCloser
	stdout io.ReadCloser
	mutex  sync.Mutex
}

/*
NewEnvironment creates a new Environment instance.
It takes a context and the stdin/stdout streams from the container.
*/
func NewEnvironment(ctx context.Context, stdin io.WriteCloser, stdout io.ReadCloser) *Environment {
	ctx, cancel := context.WithCancel(ctx)
	return &Environment{
		ctx:    ctx,
		cancel: cancel,
		stdin:  stdin,
		stdout: stdout,
	}
}

/*
Read implements io.Reader, allowing you to read from the environment's stdout.
This method is thread-safe.
*/
func (e *Environment) Read(p []byte) (n int, err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.stdout.Read(p)
}

/*
Write implements io.Writer, allowing you to write to the environment's stdin.
This method is thread-safe.
*/
func (e *Environment) Write(p []byte) (n int, err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.stdin.Write(p)
}

/*
Close closes the environment, releasing all associated resources.
This method is idempotent and thread-safe.
*/
func (e *Environment) Close() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Cancel the context to signal all operations to stop
	e.cancel()

	// Close stdin and stdout
	var err1, err2 error
	if e.stdin != nil {
		err1 = e.stdin.Close()
		e.stdin = nil
	}
	if e.stdout != nil {
		err2 = e.stdout.Close()
		e.stdout = nil
	}

	// Return the first non-nil error, if any
	if err1 != nil {
		return err1
	}
	return err2
}
