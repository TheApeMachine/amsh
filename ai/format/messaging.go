package format

type Messaging struct {
	Topic         string       `json:"topic" jsonschema_description:"The topic of the message"`
	Evaluation    []Evaluation `json:"evaluation" jsonschema_description:"Evaluation of each request, instruction, or derived action found in the message"`
	FinalResponse Reply        `json:"final_response" jsonschema_description:"Reply to the message"`
}

type Evaluation struct {
	Request string `json:"request" jsonschema_description:"The request that was made"`
	CanDo   bool   `json:"can_do" jsonschema_description:"Whether the request can be done, based on abilities and resources"`
}

type Reply struct {
	Accepted bool   `json:"accepted" jsonschema_description:"Decision to accept or reject the request"`
	Reason   string `json:"reason" jsonschema_description:"Reason for the decision"`
}
