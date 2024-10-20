package mastercomputer

import (
	"strings"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

/*
Messaging deals with updating messages, and sending them onto the queue.
The definition of a message is as follows:

- origin : the current sender of the message
- role   : the role of the message, determines the process flow
- scope  : the current intended recipient of the message (direct or topic/broadcast)
- system : the current system prompt
- user   : the current user prompt
- payload: a continuously updated log of events that took place
*/
type Messaging struct {
	worker  *Worker
	message *data.Artifact
}

func NewMessaging(worker *Worker, message *data.Artifact) *Messaging {
	return &Messaging{worker: worker, message: message}
}

/*
Process sets the worker's process flow based on the message's role.
It may only update the worker's process flow if the worker is ready.
*/
func (messaging *Messaging) Process() bool {
	// Check if the worker is ready.
	if messaging.worker.state != WorkerStateReady {

		return false
	}

	// Retrieve the process flow and assign it to the worker.
	if process := NewProcess(messaging.message.Peek("role")); process != nil {
		messaging.worker.process = process
	}

	// Add the worker's name to the process chain in the message.
	chain := strings.Split(messaging.message.Peek("chain"), ",")
	chain = append(chain, messaging.worker.name)
	messaging.message.Poke("chain", strings.Join(chain, ","))

	errnie.Info("MESSAGING: %s", messaging.message.Peek("chain"))

	// Only proceed if the worker has a process flow.
	return messaging.worker.process != nil
}

/*
Update the message according to the worker and the current process flow step.
*/
func (messaging *Messaging) Update(step map[string]string) {
	// Update the message's origin to the worker's name.
	messaging.message.Poke("origin", messaging.worker.name)

	// Update the message's role and scope according to the process flow step.
	messaging.message.Poke("role", step["role"])

	// Update the message's scope according to the process flow step.
	messaging.message.Poke("scope", step["scope"])

	// If the next scope is "previous", the message needs to go directly to the previous worker,
	// so we override the message's scope.
	if step["scope"] == "previous" {
		chain := strings.Split(messaging.message.Peek("chain"), ",")
		messaging.message.Poke("scope", chain[len(chain)-2])
	}
}
