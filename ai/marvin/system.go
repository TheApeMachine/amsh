package marvin

import (
	"context"
	"io"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

type System struct {
	pr *io.PipeReader
	pw *io.PipeWriter
}

func NewSystem() *System {
	errnie.Trace("%s", "System.NewSystem", "new")
	pr, pw := io.Pipe()

	return &System{
		pr: pr,
		pw: pw,
	}
}

func (system *System) Read(p []byte) (n int, err error) {
	n, err = system.pr.Read(p)
	if err != nil {
		return n, err
	}

	if n > 0 {
		artifact := data.Empty()
		if err := artifact.Unmarshal(p[:n]); err != nil {
			errnie.Error(err)
			// Continue even if unmarshal fails - the raw data will still be returned
		} else {
			errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))
		}
	}

	return n, nil
}

func (system *System) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		artifact := data.Empty()
		if err := artifact.Unmarshal(p); err != nil {
			errnie.Error(err)
		} else {
			errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))
		}
	}

	// Create an agent
	agent := NewAgent(context.Background(), "assistant")

	// Write to agent and copy response back to system pipe
	go func() {
		defer system.pw.Close()

		// Write to agent
		if _, err := agent.Write(p); err != nil {
			errnie.Error(err)
			return
		}

		// Copy from agent to system pipe
		buf := make([]byte, 1024)
		for {
			n, err := agent.Read(buf)
			if err != nil {
				if err != io.EOF {
					errnie.Error(err)
				}
				return
			}
			if n > 0 {
				if _, err := system.pw.Write(buf[:n]); err != nil {
					errnie.Error(err)
					return
				}
			}
		}
	}()

	return len(p), nil
}

func (system *System) Close() error {
	errnie.Trace("%s", "System.Close", "close")

	if err := system.pw.Close(); err != nil {
		return err
	}

	return system.pr.Close()
}
