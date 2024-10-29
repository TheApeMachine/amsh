package process

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/berrt"
	"github.com/theapemachine/amsh/errnie"
)

/*
Thinking is a process that allows the system to think about a given topic.
*/
type Thinking struct {
	Topic       string       `json:"topic" jsonschema:"required; description:The topic to think about"`
	Thoughts    []Thought    `json:"thoughts" jsonschema:"required; description:The thoughts about the topic"`
	Requests    []Request    `json:"requests" jsonschema:"required; description:The requests for information about the topic"`
	Conclusions []Conclusion `json:"conclusions" jsonschema:"required; description:The conclusions about the topic"`
}

/*
Thought is a single thought about a given topic.
*/
type Thought struct {
	Thought     string    `json:"thought" jsonschema:"required; description:The thought about the topic"`
	SubThoughts []Thought `json:"sub_thoughts" jsonschema:"description:The sub-thoughts about the topic"`
}

/*
Request is a single request for information about a given topic.
*/
type Request struct {
	Request string `json:"request" jsonschema:"required; description:The request for a resource"`
}

/*
Conclusion is a single conclusion about a given topic.
*/
type Conclusion struct {
	Conclusion string     `json:"conclusion" jsonschema:"required; description:The conclusion about the topic"`
	Reasoning  string     `json:"reasoning" jsonschema:"required; description:The reasoning for the conclusion"`
	Evidence   []Evidence `json:"evidence" jsonschema:"required; description:The evidence for the conclusion"`
}

/*
Evidence is a single piece of evidence supporting a conclusion.
*/
type Evidence struct {
	Fact   string `json:"fact" jsonschema:"required; description:The fact supporting the conclusion"`
	Source string `json:"source" jsonschema:"required; description:The source of the fact"`
}

/*
NewThinking creates a new instance of the Thinking process.
*/
func NewThinking() *Thinking {
	return &Thinking{}
}

/*
Marshal the process into JSON.
*/
func (thinking *Thinking) Marshal() ([]byte, error) {
	return json.Marshal(thinking)
}

/*
Unmarshal the process from JSON.
*/
func (thinking *Thinking) Unmarshal(data []byte) error {
	return json.Unmarshal(data, thinking)
}

func (thinking *Thinking) GenerateSchema() string {
	schema := jsonschema.Reflect(&Thinking{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}

	return string(out)
}

/*
Format the process as a pretty-printed JSON string.
*/
func (thinking *Thinking) Format() string {
	pretty, _ := json.MarshalIndent(thinking, "", "  ")
	return string(pretty)
}

/*
String returns a human-readable string representation of the process.
*/
func (thinking *Thinking) String() {
	berrt.Info("Thinking", thinking)
}
