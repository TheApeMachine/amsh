package mastercomputer

import (
	"context"
	"time"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/utils"
)

// Sequencer is responsible for orchestrating the execution of workers in a sequence.
// It should determine the next role to be executed based on the conversation history.
type Sequencer struct {
	ctx     context.Context
	cancel  context.CancelFunc
	workers map[string][]*Worker
	system  string
	user    string
	events  *Events
}

// NewSequencer creates a new Sequencer instance with a given message.
// The message represents the original user prompt.
func NewSequencer(ctx context.Context, message string) *Sequencer {
	sequencerCtx, cancel := context.WithCancel(ctx)

	sequencer := &Sequencer{
		ctx:     sequencerCtx,
		cancel:  cancel,
		workers: make(map[string][]*Worker),
		system:  viper.GetViper().GetString("ai.prompt.system"),
		user:    message,
		events:  NewEvents(),
	}

	// Emit an event for sequencer creation
	sequencer.events.channel <- Event{
		Timestamp: time.Now(),
		Type:      "SequencerCreated",
		Message:   "Sequencer initialized with user prompt.",
	}

	return sequencer
}

func (sequencer *Sequencer) Start() {
	// Create initial sequencer worker
	worker := sequencer.NewWorker(map[string]any{
		"role": "sequencer",
		"name": utils.NewName(),
	})
	worker.Start()

	// Send the initial user message as a task
	task := sequencer.makeTask(sequencer.system, sequencer.user, worker.role, worker.name)
	worker.task <- task

	sequencer.events.channel <- Event{
		Timestamp: time.Now(),
		Type:      "SequencerStart",
		Message:   "Initial task assigned to sequencer",
		WorkerID:  worker.name,
	}

	go func() {
		for {
			select {
			case <-sequencer.ctx.Done():
				sequencer.events.channel <- Event{
					Timestamp: time.Now(),
					Type:      "SequencerCancelled",
					Message:   "Sequencer context cancelled.",
				}
				for _, workers := range sequencer.workers {
					for _, worker := range workers {
						worker.cancel()
					}
				}

				return
			default:
				if worker := sequencer.Next(); worker != nil {
					// Let the worker process its task
					time.Sleep(100 * time.Millisecond)
				}
				sequencer.checkState()
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (sequencer *Sequencer) checkState() {
	count := 0

	for _, workers := range sequencer.workers {
		for _, worker := range workers {
			if worker.state == WorkerStateDone {
				count++
			}
		}
	}

	if count == len(sequencer.workers) {
		sequencer.events.channel <- Event{
			Timestamp: time.Now(),
			Type:      "SequencerCompleted",
			Message:   "All workers have completed their tasks.",
		}
		sequencer.cancel()
	}
}

func (sequencer *Sequencer) NewWorker(parameters map[string]any) *Worker {
	worker := NewWorker(parameters)
	sequencer.workers[parameters["role"].(string)] = append(sequencer.workers[parameters["role"].(string)], worker)

	// Emit an event for worker creation
	sequencer.events.channel <- Event{
		Timestamp: time.Now(),
		Type:      "WorkerCreated",
		Message:   "New worker added: " + worker.name,
		WorkerID:  worker.name,
	}

	return worker
}

func (sequencer *Sequencer) Next() *Worker {
	// First check if we have any active workers that aren't done
	for _, workers := range sequencer.workers {
		for _, worker := range workers {
			if worker.state == WorkerStateWorking {
				worker.task <- nil
				return worker
			}
		}
	}

	// If no workers are working, find an available sequencer to determine next steps
	for _, worker := range sequencer.workers["sequencer"] {
		if worker.state == WorkerStateReady {
			task := sequencer.makeTask(sequencer.system, sequencer.user, worker.role, worker.name)
			worker.task <- task

			sequencer.events.channel <- Event{
				Timestamp: time.Now(),
				Type:      "TaskAssigned",
				Message:   "Sequencer determining next steps",
				WorkerID:  worker.name,
			}

			return worker
		}
	}

	return nil
}

// makeContinuationTask creates a task that continues the worker's current conversation
func (sequencer *Sequencer) makeContinuationTask(worker *Worker) *Task {
	task := &Task{}
	return task
}

func (sequencer *Sequencer) makeTask(system, user, role, name string) *Task {
	return NewTask(name, role, system, user)
}
