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
			ctx, key, "layering", "manager", layering.NewProcess().SystemPrompt(key), nil,
		),
	}
}

func (pm *ProcessManager) Execute(request string) <-chan provider.Event {
	errnie.Info("Execute request %s", request)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		var layerAccumulator string

		for event := range pm.manager.Execute(request) {
			layerAccumulator += event.Content
			out <- event
		}

		accumulators := make(map[int]string)

		for idx, process := range pm.validate(layerAccumulator) {
			if toolcall := pm.checkToolCall(process); toolcall != nil {
				out <- toolcall
				continue
			}

			for _, layer := range process.Layers {
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

		errnie.Debug("process manager %s completed", pm.key)
	}()

	return out
}

func (pm *ProcessManager) validate(accumulator string) []layering.Process {
	codeBlocks := utils.ExtractCodeBlocks(accumulator)

	processes := []layering.Process{}

	if code, ok := codeBlocks["json"]; ok {
		var process layering.Process
		json.Unmarshal([]byte(code), &process)
		processes = append(processes, process)
	}

	return processes
}

func (pm *ProcessManager) checkToolCall(toolcall string) *provider.Event {
	var data map[string]any
	errnie.MustVoid(json.Unmarshal([]byte(toolcall), &data))

	if toolValue, ok := data["tool_name"].(string); ok {
		return &provider.Event{
			Type:    provider.EventToolCall,
			Content: pm.manager.Toolset.Use(pm.ctx, toolValue, data),
		}
	}

	return nil
}
