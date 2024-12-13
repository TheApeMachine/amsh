package marvin

import (
	"context"
	"io"

	"github.com/theapemachine/amsh/ai/provider"
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
	// Only try to unmarshal if we have data
	if len(p) > 0 {
		artifact := data.Empty()
		artifact.Unmarshal(p)
		errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))
	}
	return system.pr.Read(p)
}

func (system *System) Write(p []byte) (n int, err error) {
	// Only try to unmarshal if we have data
	if len(p) > 0 {
		artifact := data.Empty()
		artifact.Unmarshal(p)
		errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))
	}

	// Write to pipe in goroutine to prevent blocking
	go func() {
		defer system.pw.Close()

		// Create an agent with a balanced provider
		agent := NewAgent(context.Background(), "assistant")
		prvdr := provider.NewBalancedProvider()

		// Write the input artifact to provider through agent
		if _, err := agent.Write(p); err != nil {
			errnie.Error(err)
			return
		}

		// Copy from provider to system pipe
		buf := make([]byte, 1024)
		for {
			n, err := prvdr.Read(buf)
			if err != nil {
				if err != io.EOF {
					errnie.Error(err)
				}
				break
			}
			system.pw.Write(buf[:n])
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
