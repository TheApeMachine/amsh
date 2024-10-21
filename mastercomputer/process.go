package mastercomputer

import (
	"strconv"

	"github.com/spf13/viper"
)

var maxIterations = viper.GetViper().GetInt("ai.max_iterations")

type Process struct {
	flow []map[string]string
	log  []string
}

func NewProcess(key string) *Process {
	// Check if the process flow exists.
	if flow, ok := processMap[key]; ok {
		return &Process{flow: flow, log: make([]string, 0)}
	}

	return nil
}

var processMap = map[string][]map[string]string{
	"message": {
		{
			"role":       "reply",
			"scope":      "broadcast",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following message has been broadcasted. Please reply to the message.",
			"iterations": "1",
		},
	},
	"task": {
		{
			"role":       "execution",
			"scope":      "verifying",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following task has been assigned.",
			"iterations": strconv.Itoa(maxIterations),
		},
	},
	"execution": {
		{
			"role":       "verification",
			"scope":      "previous",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following task has been executed, please verify the effectiveness of the execution method, and provide feedback on how the worker could improve their approach in general.",
			"iterations": "1",
		},
	},
	"verification": {
		{
			"role":       "verification",
			"scope":      "finish",
			"state":      "busy",
			"done":       "ready",
			"user":       "You have received the following feedback on your execution. Please reflect on it and come up with some additions to add to your system prompt that will help you to improve your approach and performance.",
			"iterations": "1",
		},
	},
	"slack": {
		{
			"role":       "incoming",
			"scope":      "managing",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following message was posted on Slack. Please break down the message. Use your search tools to find relevant history or context.",
			"iterations": "3",
		},
	},
	"incoming": {
		{
			"role":       "delegation",
			"scope":      "dynamic",
			"state":      "busy",
			"done":       "ready",
			"user":       "The communicator has shared the following breakdown of a message received on Slack. Please provide your plan and delegate the steps to the appropriate channels.",
			"iterations": "3",
		},
	},
	"trengo": {
		{
			"role":       "helpdesk",
			"scope":      "dynamic",
			"state":      "busy",
			"done":       "ready",
			"user":       "The following came in from Trengo. Come up with the appropriate plan, make sure we extract labels matching the labels available in Trengo, and see if we need to add or update tickets in Azure.",
			"iterations": strconv.Itoa(maxIterations),
		},
	},
}
