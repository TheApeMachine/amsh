package twoface

import (
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

// Generator is a function type that processes artifacts and writes results to a channel
type Generator func(artifacts []*data.Artifact, out chan<- *data.Artifact)

// Accumulator provides a reusable generator pattern with consistent channel management
type Accumulator struct {
	buffer  *data.Artifact
	in      []*data.Artifact
	out     chan *data.Artifact
	through chan *data.Artifact
	origin  string
	role    string
	scope   string
}

// NewAccumulator creates a new Accumulator instance
func NewAccumulator(origin, role, scope string, artifacts ...*data.Artifact) *Accumulator {
	errnie.Trace("NewAccumulator", "origin", origin, "role", role, "scope", scope, "artifacts", artifacts)

	return &Accumulator{
		buffer:  data.New(origin, role, scope, []byte{}),
		in:      artifacts,
		out:     make(chan *data.Artifact),
		through: make(chan *data.Artifact),
		origin:  origin,
		role:    role,
		scope:   scope,
	}
}

// Generate starts the wrapped generator and returns a read-only channel for results
func (accumulator *Accumulator) Generate() <-chan *data.Artifact {
	errnie.Trace("Accumulator.Generate", "origin", accumulator.origin, "role", accumulator.role, "scope", accumulator.scope)

	go func() {
		defer close(accumulator.through)

		// Forward all results from the wrapped generator
		for artifact := range accumulator.out {
			accumulator.buffer.Append(artifact.Peek("payload"))
			accumulator.through <- artifact
		}
	}()

	return accumulator.through
}

// Wrap applies the generator function to process artifacts
func (accumulator *Accumulator) Yield(generator Generator) *Accumulator {
	go generator(accumulator.in, accumulator.out)
	return accumulator
}
