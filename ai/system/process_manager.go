package system

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/process/layering"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type ProcessManager struct {
	ctx     context.Context
	cancel  context.CancelFunc
	key     string
	manager *ai.Agent
}

func NewProcessManager(key, origin string) *ProcessManager {
	errnie.Info("starting process manager %s %s", key, origin)
	ctx, cancel := context.WithCancel(context.Background())

	return &ProcessManager{
		ctx:    ctx,
		cancel: cancel,
		key:    key,
		manager: ai.NewAgent(
			ctx,
			key,
			"layering",
			"manager",
			layering.NewProcess().SystemPrompt(key),
			ai.NewToolset("process"),
		),
	}
}

func (pm *ProcessManager) Execute(request string) <-chan provider.Event {
	errnie.Info("<prompt>%s</prompt>", request)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		promptEngineer := ai.NewAgent(
			pm.ctx,
			pm.key,
			"ingress",
			"prompt_engineer",
			`
			You are an expert at optimizing user prompts for maximum performance.

			<request>
			`+request+`
			</request>

			Consider the incoming request, and provide a detailed and optimized prompt, specifically crafted for Large Language Models.

			Optimize for:
			- Granularity
			- Accuracy
			- Detail
			- Avoidance of ambiguity

			There is no way to communicate with the user, so when in doubt, just use your best judgement and improve the prompt where you see fit.
			Worst-case scenario, just pass the request through unchanged.

			Output only the new prompt, nothing else.
			`,
			nil,
		)

		var prompt string

		for event := range promptEngineer.Execute(request) {
			prompt += event.Content
			out <- event
		}

		var layerAccumulator string

		for event := range pm.manager.Execute(prompt) {
			layerAccumulator += event.Content
			out <- event
		}

		var wg sync.WaitGroup

		for _, process := range pm.validate(layerAccumulator) {
			wg.Add(2)

			// Execute the main layer branch.
			go func(process layering.Process, wg *sync.WaitGroup) {
				defer wg.Done()
				pm.processLayers(process.Layers, out)
			}(process, &wg)

			// Execute the fork branches.
			for _, fork := range process.Forks {
				go func(fork layering.Fork, wg *sync.WaitGroup) {
					defer wg.Done()
					pm.processLayers(fork.Layers, out)
				}(fork, &wg)
			}
		}

		wg.Wait()
		errnie.Debug("process manager %s completed", pm.key)
	}()

	return out
}

func (pm *ProcessManager) processLayers(layers []layering.Layer, out chan<- provider.Event) {
	accumulators := make(map[int]string)

	for idx, layer := range layers {
		errnie.Info("executing layer %s", layer.Workloads)

		var wg sync.WaitGroup
		wg.Add(len(layer.Workloads))

		ctx, cancel := context.WithCancel(pm.ctx)
		defer cancel()

		for event := range NewProcessor(ctx, pm.key).Process(layer) {
			accumulators[idx] += event.Content
			out <- event
		}

		wg.Wait()
	}
}

func (pm *ProcessManager) validate(accumulator string) []layering.Process {
	errnie.Log("validating process manager %s", accumulator)
	codeBlocks := utils.ExtractCodeBlocks(accumulator)
	errnie.Log("code blocks %v", codeBlocks)

	processes := []layering.Process{}

	if blocks, ok := codeBlocks["json"]; ok {
		errnie.Log("json blocks %v", blocks)
		for _, code := range blocks {
			errnie.Log("code %s", code)
			if pm.checkToolCall(code) {
				continue
			}

			var process layering.Process
			errnie.MustVoid(json.Unmarshal([]byte(code), &process))
			processes = append(processes, process)
		}
	}

	return processes
}

func (pm *ProcessManager) checkToolCall(toolcall string) bool {
	var data map[string]any
	errnie.MustVoid(json.Unmarshal([]byte(toolcall), &data))

	if toolValue, ok := data["tool_name"].(string); ok {
		errnie.Info("executing tool call %s - %v", toolValue, data)
		pm.manager.Toolset.Use(pm.ctx, toolValue, data)
		return true
	}

	return false
}
