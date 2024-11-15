package mastercomputer

import (
	"context"
	"fmt"
	"time"

	"github.com/theapemachine/amsh/ai/process/quantum"
	"github.com/theapemachine/amsh/ai/process/temporal"
	"github.com/theapemachine/amsh/ai/process/tensor"
	"github.com/theapemachine/qpool"
)

var (
	ErrBack   = fmt.Errorf("operation requests back")
	ErrCancel = fmt.Errorf("operation cancelled")
)

type Operation struct {
	Type       string
	Behavior   string
	Parameters map[string]interface{}
}

/*
Processor runs compiled instructions by spawning worker agents
for each operation. It maintains the context state and handles
the flow of data between operations.

Context travels:
- Bottom-to-top, right-to-left between closures
- Top-to-bottom, left-to-right within closures
*/
type Processor struct {
	ctx        context.Context
	agentPool  *qpool.Q
	contextBus chan Context // Manages context flow between operations
	errBus     chan error   // Error channel for operation failures
}

// Context represents the mutable state passed between operations
type Context struct {
	Data      interface{}
	Mutations []string // Track operations that modified this context
	Parent    *Context // For hierarchical context tracking
}

func NewProcessor(ctx context.Context) *Processor {
	// Create qpool with reasonable defaults
	pool := qpool.NewQ(ctx, 5, 20, &qpool.Config{
		SchedulingTimeout: time.Minute,
	})

	return &Processor{
		ctx:        ctx,
		agentPool:  pool,
		contextBus: make(chan Context, 100), // Buffer for context flow
		errBus:     make(chan error, 100),   // Buffer for errors
	}
}

/*
Execute runs a sequence of instructions by:
1. Spawning worker agents for operations
2. Managing context flow between operations
3. Handling concurrent execution in closures
*/
func (proc *Processor) Execute(instructions []Instruction) (Context, error) {
	var currentContext Context

	for i := 0; i < len(instructions); i++ {
		inst := instructions[i]

		var err error
		switch inst.Op {
		case OpCall:
			currentContext, i, err = proc.executeOperation(inst, i, instructions, currentContext)
		case OpJoin:
			currentContext, err = proc.executeJoin(inst, currentContext)
		}

		if err != nil {
			return currentContext, err
		}
	}

	return currentContext, nil
}

func (proc *Processor) executeOperation(inst Instruction, i int, instructions []Instruction, currentContext Context) (Context, int, error) {
	result := proc.agentPool.Schedule(
		fmt.Sprintf("op-%d", i),
		func() (any, error) {
			return proc.processOperation(inst.Operands[0].(Operation), currentContext)
		},
		qpool.WithTTL(time.Minute),
	)

	select {
	case quantumValue := <-result:
		if quantumValue.Error != nil {
			if quantumValue.Error == ErrBack {
				return currentContext, proc.findPreviousOperation(i, instructions), nil
			}
			if quantumValue.Error == ErrCancel {
				return currentContext, i, quantumValue.Error
			}
			proc.errBus <- fmt.Errorf("operation failed: %w", quantumValue.Error)
			return currentContext, i, <-proc.errBus
		}
		return quantumValue.Value.(Context), i, nil
	case <-proc.ctx.Done():
		return currentContext, i, proc.ctx.Err()
	}
}

func (proc *Processor) processOperation(op Operation, currentContext Context) (Context, error) {
	var process interface{}
	switch op.Behavior {
	case "temporal":
		process = temporal.NewProcess()
	case "quantum":
		process = quantum.NewState()
	case "tensor":
		process = tensor.NewProcess()
	default:
		return Context{}, fmt.Errorf("unknown behavior type: %s", op.Behavior)
	}

	currentContext.Data = process
	currentContext.Mutations = append(currentContext.Mutations, op.Type)
	return currentContext, nil
}

func (proc *Processor) executeJoin(inst Instruction, currentContext Context) (Context, error) {
	results, err := proc.joinContexts(inst.Operands[0].(int))
	if err != nil {
		return currentContext, err
	}
	return proc.mergeContexts(results), nil
}

/*
joinContexts collects results from concurrent operations
and merges them into a single context.
*/
func (proc *Processor) joinContexts(count int) ([]Context, error) {
	contexts := make([]Context, 0, count)

	// Collect contexts from concurrent operations
	for i := 0; i < count; i++ {
		select {
		case ctx := <-proc.contextBus:
			contexts = append(contexts, ctx)
		case err := <-proc.errBus:
			return nil, err
		case <-proc.ctx.Done():
			return nil, proc.ctx.Err()
		}
	}

	return contexts, nil
}

/*
mergeContexts combines multiple contexts into a single context,
preserving the mutation history and relationships.
*/
func (proc *Processor) mergeContexts(contexts []Context) Context {
	merged := Context{
		Mutations: make([]string, 0),
	}

	// Merge data and mutations from all contexts
	for _, ctx := range contexts {
		// Merge logic depends on data type and operation requirements
		merged.Data = proc.mergeData(merged.Data, ctx.Data)
		merged.Mutations = append(merged.Mutations, ctx.Mutations...)
	}

	return merged
}

// Add WorkerConfig type
type WorkerConfig struct {
	Type     string
	Behavior string
	Params   map[string]interface{}
}

/*
findPreviousOperation locates the previous operation boundary
when handling a "back" request.
*/
func (proc *Processor) findPreviousOperation(current int, instructions []Instruction) int {
	// Look for the previous operation boundary
	for i := current - 1; i >= 0; i-- {
		// Check for operation boundaries (e.g., start of a block)
		switch instructions[i].Op {
		case OpCall, OpJoin:
			return i
		}
	}
	return 0
}

/*
mergeData combines two data objects based on their types and
the operation requirements.
*/
func (proc *Processor) mergeData(existing, new interface{}) interface{} {
	// If no existing data, use the new data
	if existing == nil {
		return new
	}

	// Implement your specific merge logic here
	// This is just a placeholder that returns the newer data
	return new
}
