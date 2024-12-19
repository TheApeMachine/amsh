package twoface

import (
	"github.com/theapemachine/amsh/data"
)

// Generator is a function type that processes artifacts and writes results to a channel
type Generator func(*Accumulator)

// Accumulator provides a reusable generator pattern with consistent channel management
type Accumulator struct {
	buffer  *data.Artifact
	In      []*data.Artifact
	Out     chan *data.Artifact
	through chan *data.Artifact
	origin  string
	role    string
	scope   string
}

// NewAccumulator creates a new Accumulator instance
func NewAccumulator(origin, role, scope string, artifacts ...*data.Artifact) *Accumulator {
	return &Accumulator{
		buffer:  data.New(origin, role, scope, []byte{}),
		In:      artifacts,
		Out:     make(chan *data.Artifact),
		through: make(chan *data.Artifact),
		origin:  origin,
		role:    role,
		scope:   scope,
	}
}

// Generate starts the wrapped generator and returns a read-only channel for results
func (accumulator *Accumulator) Generate() <-chan *data.Artifact {
	go func() {
		defer close(accumulator.through)

		// Clear the buffer.
		accumulator.buffer.Poke("payload", "")

		// Forward all results from the wrapped generator
		for artifact := range accumulator.Out {
			accumulator.buffer.Append(artifact.Peek("payload"))
			accumulator.through <- artifact
		}
	}()

	return accumulator.through
}

// Wrap applies the generator function to process artifacts
func (accumulator *Accumulator) Yield(generator Generator) *Accumulator {
	go generator(accumulator)
	return accumulator
}

func (accumulator *Accumulator) Take() *data.Artifact {
	return accumulator.buffer
}
