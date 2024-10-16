package format

import (
	"fmt"
	"strings"
)

// Messaging represents the structure of a messaging format.
type Messaging struct {
	Topic         string       `json:"topic" jsonschema_description:"The topic of the message"`
	Evaluation    []Evaluation `json:"evaluation" jsonschema_description:"Evaluation of each request, instruction, or derived action found in the message"`
	FinalResponse Reply        `json:"final_response" jsonschema_description:"Reply to the message"`
}

func (m Messaging) Format() ResponseFormat {
	return m
}

func (m Messaging) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Topic: %s\n", m.Topic))
	for _, evaluation := range m.Evaluation {
		sb.WriteString(fmt.Sprintf("Request: %s, Can Do: %v\n", evaluation.Request, evaluation.CanDo))
	}
	sb.WriteString(fmt.Sprintf("Accepted: %v, Reason: %s", m.FinalResponse.Accepted, m.FinalResponse.Reason))
	return sb.String()
}

type Evaluation struct {
	Request string `json:"request" jsonschema_description:"The request that was made"`
	CanDo   bool   `json:"can_do" jsonschema_description:"Whether the request can be done, based on abilities and resources"`
}

type Reply struct {
	Accepted bool   `json:"accepted" jsonschema_description:"Decision to accept or reject the request"`
	Reason   string `json:"reason" jsonschema_description:"Reason for the decision"`
}
