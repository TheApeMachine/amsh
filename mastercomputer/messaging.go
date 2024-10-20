package mastercomputer

import (
	"fmt"

	"github.com/theapemachine/amsh/data"
)

/*
Messaging deals with updating messages, and sending them onto the queue.
The definition of a message is as follows:

- origin : the current sender of the message
- role   : the role of the sender
- scope  : the current intended recipient of the message (direct or topic/broadcast)
- system : the current system prompt
- user   : the current user prompt
- payload: a continously updated log of events that took place
- chain  : an ordered list of workers that have processed the message
- stage  : the stage of the process we are in
*/
type Messaging struct {
	worker  *Worker
	message *data.Artifact
}

func NewMessaging(worker *Worker, message *data.Artifact) *Messaging {
	return &Messaging{worker: worker, message: message}
}

func (messaging *Messaging) Process() *data.Artifact {
	// Updating the chain keeps track of the workers that have processed the message.
	messaging.message.Poke("chain", fmt.Sprintf(
		"%s, %s", messaging.message.Peek("chain"), messaging.worker.name,
	))

	stateContext := scopeMap[messaging.message.Peek("role")][messaging.message.Peek("stage")]

	if stateContext["scope"] != "" {
		messaging.message.Poke("scope", stateContext["scope"])
	}

	// Prepare the message for the execution step.
	for _, key := range []string{"origin", "role", "system", "user"} {
		messaging.message.Poke(key, messaging.worker.buffer.Peek(key))
	}

	messaging.worker.NewState(
		messaging.worker.StateByKey(stateContext["state"]),
	)

	messaging.message.Poke("stage", stateContext["stage"])
	return messaging.message
}

var scopeMap = map[string]map[string]map[string]string{
	"manager": {
		"ingress": {
			"scope": "verifier",
			"state": "busy",
			"stage": "executing",
		},
		"executing": {
			"scope": "previous",
			"state": "busy",
			"stage": "verifying",
		},
	},
}
