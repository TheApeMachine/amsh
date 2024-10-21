package mastercomputer

import (
	"strconv"

	"github.com/spf13/viper"
)

var maxIterations = strconv.Itoa(viper.GetViper().GetInt("ai.max_iterations"))

type Process struct {
	flow []map[string]string
	log  []string
}

/*
NewProcess returns the matching process flow for the current state of the message, or nil if no flow is found.
When the process is returned as nil, the message is discarded by the current worker.
*/
func NewProcess(key string) *Process {
	// Check if the process flow exists.
	if flow, ok := processMap[key]; ok {
		return &Process{flow: flow, log: make([]string, 0)}
	}

	return nil
}

/*
processMap contains process flows that keep the system on track, preventing it collapsing into chaos.
Essentially it contains the values that are used to update a message, which guides the Worker on how to act,
or navigates the system through conditional logic.

DEFINITIONS:
- key: matches the current role of the message.
  - role      : sets the next role of the message, which will determine which key is going to be used by the next worker.
  - scope     : sets the next scope of the message, which will determine the next recipient, topic channel, or dynamic value..
  - state     : the state to put the worker in while it is processing/executing the message.
  - done      : the state to set the message to after the worker has finished processing/executing it.
  - user      : the user prompt to set on the message, which acts as a dynamic additional instruction, and will be merged with the worker's user prompt, and the message payload.
  - iterations: the maximum number of iterations the worker that should handle the message is allowed to perform.
*/
var processMap = map[string][]map[string]string{
	"message": {
		{
			"role":       "reply",
			"scope":      "broadcast",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following message has been broadcasted. Please reply to the message.",
			"iterations": maxIterations,
		},
	},
	"reply": {
		{
			"role":       "reply",
			"scope":      "previous",
			"state":      "busy",
			"done":       "ready",
			"user":       "You have received a reply.",
			"iterations": maxIterations,
		},
	},
	"task": {
		{
			"role":       "execution",
			"scope":      "verifying",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following task has been assigned.",
			"iterations": maxIterations,
		},
	},
	"execution": {
		{
			"role":       "verification",
			"scope":      "previous",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following task has been executed, please verify the effectiveness of the execution method, and provide feedback on how the worker could improve their approach in general.",
			"iterations": maxIterations,
		},
	},
	"verification": {
		{
			"role":       "verification",
			"scope":      "finish",
			"state":      "busy",
			"done":       "ready",
			"user":       "You have received the following feedback on your execution. Please reflect on it and come up with some additions to add to your system prompt that will help you to improve your approach and performance.",
			"iterations": maxIterations,
		},
	},
	"slack": {
		{
			"role":       "incoming",
			"scope":      "managing",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following message was posted on Slack. Please break down the message. Use your search tools to find relevant history or context.",
			"iterations": maxIterations,
		},
	},
	"incoming": {
		{
			"role":       "delegation",
			"scope":      "dynamic",
			"state":      "busy",
			"done":       "ready",
			"user":       "The communicator has shared the following breakdown of a message received on Slack. Please provide your plan and delegate the steps to the appropriate channels.",
			"iterations": maxIterations,
		},
	},
	"trengo": {
		{
			"role":       "helpdesk",
			"scope":      "dynamic",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following came in from Trengo. Come up with the appropriate plan, make sure we extract labels matching the labels available in Trengo, and see if we need to add or update tickets in Azure.",
			"iterations": maxIterations,
		},
	},
	"github": {
		{
			"role":       "delegation",
			"scope":      "dynamic",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following came in from GitHub. Please break down the message and delegate the steps to the appropriate channels.",
			"iterations": maxIterations,
		},
	},
}
