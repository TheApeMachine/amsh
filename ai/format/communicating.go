package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Communicating struct {
	Breakdown Breakdown `json:"breakdown" jsonschema:"description=A breakdown of the message"`
	Done      bool      `json:"done" jsonschema:"description=Indicates whether you are done thinking about the message and ready with your breakdown"`
}

func NewCommunicating() *Communicating {
	return &Communicating{}
}

func (communicating *Communicating) Print(data []byte) (isDone bool, err error) {
	if err := errnie.Error(json.Unmarshal(data, communicating)); err != nil {
		return communicating.Done, err
	}

	fmt.Println(communicating.String())
	return communicating.Done, nil
}

func (cp Communicating) String() string {
	output := utils.Dark("[COMMUNICATION]") + "\n"
	output += "\t" + utils.Muted("[BREAKDOWN]") + "\n"
	output += "\t\t" + utils.Blue("Message Type: ") + cp.Breakdown.MessageType + "\n"
	output += "\t\t" + utils.Green("Intent: ") + cp.Breakdown.Intent + "\n"
	output += "\t\t" + utils.Yellow("Sentiment: ") + cp.Breakdown.Sentiment + "\n"
	output += "\t\t" + utils.Muted("[CONTEXT]") + "\n"
	output += "\t\t\t" + utils.Red("Sender: ") + cp.Breakdown.MessageContext.Sender + "\n"
	output += "\t\t\t" + utils.Red("Channel: ") + cp.Breakdown.MessageContext.Channel + "\n"
	output += "\t\t\t" + utils.Red("Time: ") + cp.Breakdown.MessageContext.Time + "\n"
	output += "\t\t" + utils.Muted("[/CONTEXT]") + "\n"
	output += "\t\t" + utils.Muted("[ENTITIES]") + "\n"
	output += "\t\t\t" + utils.Red("People: ") + strings.Join(cp.Breakdown.Entities.People, ", ") + "\n"
	output += "\t\t\t" + utils.Red("Objects: ") + strings.Join(cp.Breakdown.Entities.Objects, ", ") + "\n"
	output += "\t\t\t" + utils.Red("Dates: ") + strings.Join(cp.Breakdown.Entities.Dates, ", ") + "\n"
	output += "\t\t" + utils.Muted("[/ENTITIES]") + "\n"
	output += "\t\t" + utils.Blue("Keywords: ") + strings.Join(cp.Breakdown.Keywords, ", ") + "\n"
	output += "\t" + utils.Muted("[/BREAKDOWN]") + "\n"
	output += "\t" + utils.Blue("Done: ") + BoolToString(cp.Done) + "\n"
	output += utils.Dark("[/COMMUNICATION]")
	return output
}

type Breakdown struct {
	MessageType     string   `json:"message_type" jsonschema:"description=The type of message (Question, Command, Statement, Acknowledgment, etc.)"`
	MessageContext  Context  `json:"context" jsonschema:"description=Details about who sent the message, where, and when"`
	Intent          string   `json:"intent" jsonschema:"description=The implied purpose of the message (Actionable, Informational, Urgent, etc.)"`
	Objectives      []string `json:"objectives" jsonschema:"description=Any significant objectives or goals mentioned in the message"`
	Entities        Entities `json:"entities" jsonschema:"description=People, objects, or dates mentioned in the message"`
	Sentiment       string   `json:"sentiment" jsonschema:"description=The emotional tone of the message"`
	RelevantHistory []string `json:"relevant_history" jsonschema:"description=Any relevant history or context that helps understand the message"`
	Keywords        []string `json:"keywords" jsonschema:"description=Key topics or important terms in the message"`
}

type Context struct {
	Sender  string `json:"sender" jsonschema:"description=The person or entity that sent the message"`
	Channel string `json:"channel" jsonschema:"description=Where the message was sent"`
	Time    string `json:"time" jsonschema:"description=Timestamp when the message was sent"`
}

type Entities struct {
	People   []string `json:"people" jsonschema:"description=People mentioned in the message"`
	Places   []string `json:"places" jsonschema:"description=Places mentioned in the message"`
	Products []string `json:"products" jsonschema:"description=Products mentioned in the message"`
	Projects []string `json:"projects" jsonschema:"description=Projects mentioned in the message"`
	Objects  []string `json:"objects" jsonschema:"description=Items or things referred to in the message"`
	Dates    []string `json:"dates" jsonschema:"description=Any dates or times referenced in the message"`
}
