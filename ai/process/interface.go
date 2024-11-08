package process

import (
	"github.com/theapemachine/amsh/ai/process/context"
	"github.com/theapemachine/amsh/ai/process/development"
	"github.com/theapemachine/amsh/ai/process/fractal"
	"github.com/theapemachine/amsh/ai/process/graph"
	"github.com/theapemachine/amsh/ai/process/holographic"
	"github.com/theapemachine/amsh/ai/process/ideation"
	"github.com/theapemachine/amsh/ai/process/quantum"
	"github.com/theapemachine/amsh/ai/process/research"
	"github.com/theapemachine/amsh/ai/process/story"
	"github.com/theapemachine/amsh/ai/process/temporal"
	"github.com/theapemachine/amsh/ai/process/tensor"
)

/*
Process defines an interface that object can implement if the want to act
as a predefined process. Predefined processes are used to direct specific
behavior, useful is cases where we know what should be done based on an input.
*/
type Process interface {
	GenerateSchema() string
}

/*
ProcessMap finds a single process by key, which is used to map incoming
WebHooks to the correct pre-defined process.
*/
var ProcessMap = map[string]Process{
	"temporal_dynamics":  &temporal.Process{},
	"holographic_memory": &holographic.Process{},
	"fractal_structure":  &fractal.Process{},
	"hypergraph":         &graph.Process{},
	"tensor_network":     &tensor.Process{},
	"quantum_layer":      &quantum.Process{},
	"ideation":           &ideation.Process{},
	"context_mapping":    &context.Process{},
	"story_flow":         &story.Process{},
	"research":           &research.Process{},
	"architecture":       &development.Architecture{},
	"requirements":       &development.Requirements{},
	"implementation":     &development.Implementation{},
	"testing":            &development.Testing{},
	"deployment":         &development.Deployment{},
	"documentation":      &development.Documentation{},
	"review":             &development.Review{},
}
