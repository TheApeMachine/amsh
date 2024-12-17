package twoface

import (
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

type Generator func(artifacts []*data.Artifact, out chan<- *data.Artifact, accumulator *Accumulator) *Accumulator

type Accumulator struct {
	in     []*data.Artifact
	out    chan *data.Artifact
	origin string
	role   string
	scope  string
}

func NewAccumulator(origin, role, scope string, artifacts ...*data.Artifact) *Accumulator {
	errnie.Trace("NewAccumulator", "origin", origin, "role", role, "scope", scope, "artifacts", artifacts)

	return &Accumulator{
		in:     artifacts,
		out:    make(chan *data.Artifact),
		origin: origin,
		role:   role,
		scope:  scope,
	}
}

func (accumulator *Accumulator) Generate() <-chan *data.Artifact {
	errnie.Trace("Accumulator.Generate", "origin", accumulator.origin, "role", accumulator.role, "scope", accumulator.scope)
	out := make(chan *data.Artifact)

	go func() {
		defer close(accumulator.out)
		defer close(out)

		// Then stream any results from wrapped generators
		for artifact := range accumulator.out {
			errnie.Trace("Accumulator.Generate", "status", "Passing through wrapped generator artifact", "artifact_payload", artifact.Peek("payload"))
			out <- artifact
		}
	}()
	errnie.Trace("Accumulator.Generate", "status", "Generator started")
	return out
}

func (accumulator *Accumulator) Wrap(generator Generator) *Accumulator {
	errnie.Trace("Wrap", "generator", generator)

	return generator(accumulator.in, accumulator.out, accumulator)
}
