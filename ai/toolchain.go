package ai

import (
	"context"
	"sync"
	"time"
)

// ExecutionMode determines how steps are executed
type ExecutionMode string

type ToolResult struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

const (
	Sequential ExecutionMode = "sequential"
	Parallel   ExecutionMode = "parallel"
	Pipeline   ExecutionMode = "pipeline" // Like parallel but with data streaming
)

// ErrorStrategy determines how to handle errors
type ErrorStrategy string

const (
	StopOnError     ErrorStrategy = "stop"
	ContinueOnError ErrorStrategy = "continue"
	RetryOnError    ErrorStrategy = "retry"
	Fallback        ErrorStrategy = "fallback"
)

// ToolChain represents a sequence of tool operations
type ToolChain struct {
	Steps         []ToolStep    `json:"steps" jsonschema:"required,description=Sequence of tool operations to perform"`
	Mode          ExecutionMode `json:"mode" jsonschema:"required,description=How to execute the steps"`
	ErrorStrategy ErrorStrategy `json:"error_strategy" jsonschema:"description=How to handle errors"`
	MaxRetries    int           `json:"max_retries,omitempty" jsonschema:"description=Maximum number of retries for failed steps"`
	RetryDelay    string        `json:"retry_delay,omitempty" jsonschema:"description=Delay between retries (e.g., '1s', '100ms')"`
	Timeout       string        `json:"timeout,omitempty" jsonschema:"description=Maximum execution time for the chain"`
}

// ToolStep enhanced with error handling and execution options
type ToolStep struct {
	Tool          string         `json:"tool" jsonschema:"required,description=Name of the tool to use"`
	Args          map[string]any `json:"args" jsonschema:"required,description=Arguments for the tool"`
	OutputVar     string         `json:"output_var,omitempty" jsonschema:"description=Variable name to store the output"`
	InputVars     []string       `json:"input_vars,omitempty" jsonschema:"description=Variables to use from previous steps"`
	Condition     string         `json:"condition,omitempty" jsonschema:"description=Condition for executing this step"`
	ErrorStrategy ErrorStrategy  `json:"error_strategy,omitempty" jsonschema:"description=Override chain's error strategy"`
	Fallback      *ToolStep      `json:"fallback,omitempty" jsonschema:"description=Fallback step if this one fails"`
	Timeout       string         `json:"timeout,omitempty" jsonschema:"description=Maximum execution time for this step"`
	Weight        int            `json:"weight,omitempty" jsonschema:"description=Execution priority for parallel processing"`
}

// StepResult represents the result of executing a single step
type StepResult struct {
	StepIndex int
	Result    ToolResult
	Retries   int
	Duration  time.Duration
	Error     error
}

// ToolPipeline enhanced with advanced execution features
type ToolPipeline struct {
	toolset      *Toolset
	vars         sync.Map
	results      []StepResult
	ctx          context.Context
	errorHandler ErrorHandler
}

// ErrorHandler manages error recovery strategies
type ErrorHandler struct {
	strategy   ErrorStrategy
	maxRetries int
	retryDelay time.Duration
}

func NewToolPipeline(ctx context.Context, toolset *Toolset) *ToolPipeline {
	return &ToolPipeline{
		toolset: toolset,
		ctx:     ctx,
	}
}

// Execute runs a tool chain with the specified execution mode
func (tp *ToolPipeline) Execute(chain ToolChain) ToolResult {
	// Set up error handling
	tp.errorHandler = ErrorHandler{
		strategy:   chain.ErrorStrategy,
		maxRetries: chain.MaxRetries,
	}
	if chain.RetryDelay != "" {
		if delay, err := time.ParseDuration(chain.RetryDelay); err == nil {
			tp.errorHandler.retryDelay = delay
		}
	}

	// Set up timeout if specified
	if chain.Timeout != "" {
		if timeout, err := time.ParseDuration(chain.Timeout); err == nil {
			var cancel context.CancelFunc
			tp.ctx, cancel = context.WithTimeout(tp.ctx, timeout)
			defer cancel()
		}
	}

	// Execute based on mode
	switch chain.Mode {
	case Parallel:
		return tp.executeParallel(chain)
	case Pipeline:
		return tp.executePipeline(chain)
	default:
		return tp.executeSequential(chain)
	}
}

// executeParallel runs steps concurrently with dependency management
func (tp *ToolPipeline) executeParallel(chain ToolChain) ToolResult {
	var wg sync.WaitGroup
	results := make(chan StepResult, len(chain.Steps))
	dependencies := tp.analyzeDependencies(chain.Steps)

	// Create worker pool
	workers := make(chan struct{}, 10) // Limit concurrent executions

	// Track completed steps for dependency resolution
	completed := sync.Map{}

	for i, step := range chain.Steps {
		if !tp.canExecuteStep(i, dependencies, &completed) {
			continue
		}

		wg.Add(1)
		workers <- struct{}{} // Acquire worker

		go func(i int, step ToolStep) {
			defer wg.Done()
			defer func() { <-workers }() // Release worker

			result := tp.executeStep(i, step)
			results <- result

			if result.Result.Success {
				completed.Store(i, true)
				// Check if any pending steps can now execute
				tp.checkPendingSteps(chain, dependencies, &completed)
			}
		}(i, step)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	success := true
	finalData := make(map[string]any)
	for result := range results {
		if !result.Result.Success {
			success = false
			if chain.ErrorStrategy == StopOnError {
				break
			}
		}
		// Merge successful results
		if data, ok := result.Result.Data.(map[string]any); ok {
			for k, v := range data {
				finalData[k] = v
			}
		}
	}

	return ToolResult{
		Success: success,
		Data:    finalData,
	}
}

// executePipeline runs steps in a streaming pipeline
func (tp *ToolPipeline) executePipeline(chain ToolChain) ToolResult {
	streams := make([]chan any, len(chain.Steps)+1)
	for i := range streams {
		streams[i] = make(chan any, 1)
	}

	var wg sync.WaitGroup
	wg.Add(len(chain.Steps))

	// Start each step in the pipeline
	for i, step := range chain.Steps {
		go func(i int, step ToolStep) {
			defer wg.Done()
			defer close(streams[i+1])

			for data := range streams[i] {
				// Update args with streamed data
				if step.Args == nil {
					step.Args = make(map[string]any)
				}
				step.Args["stream_data"] = data

				result := tp.executeStep(i, step)
				if result.Result.Success {
					streams[i+1] <- result.Result.Data
				}
			}
		}(i, step)
	}

	// Feed initial data into the pipeline
	close(streams[0])

	// Wait for pipeline completion
	wg.Wait()

	return ToolResult{
		Success: true,
		Data:    "Pipeline completed",
	}
}

// executeStep with retry and fallback logic
func (tp *ToolPipeline) executeStep(index int, step ToolStep) StepResult {
	startTime := time.Now()
	result := StepResult{StepIndex: index}

	// Set up step timeout if specified
	stepCtx := tp.ctx
	if step.Timeout != "" {
		if timeout, err := time.ParseDuration(step.Timeout); err == nil {
			var cancel context.CancelFunc
			stepCtx, cancel = context.WithTimeout(tp.ctx, timeout)
			defer cancel()
		}
	}

	// Determine error strategy
	strategy := tp.errorHandler.strategy
	if step.ErrorStrategy != "" {
		strategy = step.ErrorStrategy
	}

	// Execute with retries if configured
	for retry := 0; retry <= tp.errorHandler.maxRetries; retry++ {
		result.Result = ToolResult{
			Success: true,
			Data:    tp.toolset.Use(stepCtx, step.Tool, step.Args),
		}
		result.Retries = retry

		if result.Result.Success {
			break
		}

		if retry < tp.errorHandler.maxRetries {
			time.Sleep(tp.errorHandler.retryDelay)
			continue
		}

		// If retries exhausted, try fallback
		if strategy == Fallback && step.Fallback != nil {
			fallbackResult := tp.executeStep(index, *step.Fallback)
			if fallbackResult.Result.Success {
				result = fallbackResult
				break
			}
		}
	}

	result.Duration = time.Since(startTime)
	return result
}

// analyzeDependencies builds a dependency graph for parallel execution
func (tp *ToolPipeline) analyzeDependencies(steps []ToolStep) map[int][]int {
	deps := make(map[int][]int)

	for i, step := range steps {
		deps[i] = []int{}

		// Add dependencies based on InputVars
		for _, inputVar := range step.InputVars {
			// Find step that produces this variable
			for j, prevStep := range steps[:i] {
				if prevStep.OutputVar == inputVar {
					deps[i] = append(deps[i], j)
				}
			}
		}
	}

	return deps
}

// canExecuteStep checks if all dependencies are satisfied
func (tp *ToolPipeline) canExecuteStep(stepIndex int, deps map[int][]int, completed *sync.Map) bool {
	for _, dep := range deps[stepIndex] {
		if _, ok := completed.Load(dep); !ok {
			return false
		}
	}
	return true
}

// Define ComposedTool to represent a tool with a name, description, and a tool chain
type ComposedTool struct {
	Name        string    `json:"name" jsonschema:"required,description=Name of the composed tool"`
	Description string    `json:"description" jsonschema:"description=Description of the composed tool"`
	Chain       ToolChain `json:"chain" jsonschema:"required,description=Tool chain to execute"`
}

// Example usage showing advanced features
func createAdvancedScraper() *ComposedTool {
	return &ComposedTool{
		Name:        "advanced_scraper",
		Description: "Advanced web scraping with parallel processing and error recovery",
		Chain: ToolChain{
			Mode:          Parallel,
			ErrorStrategy: RetryOnError,
			MaxRetries:    3,
			RetryDelay:    "1s",
			Timeout:       "30s",
			Steps: []ToolStep{
				{
					Tool: "browser",
					Args: map[string]any{
						"url":      "$url",
						"selector": ".content",
					},
					OutputVar: "content",
					Timeout:   "5s",
					Fallback: &ToolStep{
						Tool: "browser",
						Args: map[string]any{
							"url":  "$url",
							"mode": "simplified",
						},
					},
				},
				{
					Tool: "memory",
					Args: map[string]any{
						"operation": "store",
						"data":      "$content",
					},
					InputVars:     []string{"content"},
					ErrorStrategy: Fallback,
				},
				{
					Tool: "browser",
					Args: map[string]any{
						"url":    "$url",
						"action": "screenshot",
					},
					OutputVar: "screenshot",
					Weight:    2, // Higher priority
				},
			},
		},
	}
}

// executeSequential runs steps in a sequential manner
func (tp *ToolPipeline) executeSequential(chain ToolChain) ToolResult {
	finalData := make(map[string]any)
	success := true

	for i, step := range chain.Steps {
		result := tp.executeStep(i, step)
		if !result.Result.Success {
			success = false
			if chain.ErrorStrategy == StopOnError {
				break
			}
		}
		// Merge successful results
		if data, ok := result.Result.Data.(map[string]any); ok {
			for k, v := range data {
				finalData[k] = v
			}
		}
	}

	return ToolResult{
		Success: success,
		Data:    finalData,
	}
}

// checkPendingSteps evaluates pending steps and triggers execution if dependencies are met
func (tp *ToolPipeline) checkPendingSteps(chain ToolChain, dependencies map[int][]int, completed *sync.Map) {
	for stepIndex := range dependencies {
		if tp.canExecuteStep(stepIndex, dependencies, completed) {
			// Trigger execution of the step if all dependencies are satisfied
			tp.executeStep(stepIndex, chain.Steps[stepIndex])
		}
	}
}
