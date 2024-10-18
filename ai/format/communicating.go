package format

import (
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Communicating struct {
	InternalMessages []Message      `json:"internal_messages" jsonschema:"description=Messages for internal team members"`
	ExternalMessages []SlackMessage `json:"external_messages" jsonschema:"description=Messages for external parties via Slack"`
	Responses        []Response     `json:"responses" jsonschema:"description=Responses to incoming messages"`
	Subscriptions    []Subscription `json:"subscriptions" jsonschema:"description=Topic channels subscribed to"`
	PendingActions   []Action       `json:"pending_actions" jsonschema:"description=Communication actions to be taken"`
	Done             bool           `json:"done" jsonschema:"description=Indicates if the communication plan is complete;required=true"`
}

func NewCommunicating() *Communicating {
	return &Communicating{}
}

func (communicating *Communicating) Print(data []byte) error {
	if err := errnie.Error(json.Unmarshal(data, communicating)); err != nil {
		return err
	}

	fmt.Println(communicating.String())
	return nil
}

func (cp Communicating) String() string {
	output := utils.Dark("[COMMUNICATION PLAN]") + "\n"

	output += "\t" + utils.Muted("[INTERNAL MESSAGES]") + "\n"
	for _, msg := range cp.InternalMessages {
		output += "\t\t" + utils.Blue("To: ") + msg.Recipient + "\n"
		output += "\t\t" + utils.Green("Topic: ") + msg.Topic + "\n"
		output += "\t\t" + utils.Yellow("Message: ") + msg.Content + "\n"
	}
	output += "\t" + utils.Muted("[/INTERNAL MESSAGES]") + "\n"

	output += "\t" + utils.Muted("[EXTERNAL MESSAGES (SLACK)]") + "\n"
	for _, msg := range cp.ExternalMessages {
		output += "\t\t" + utils.Blue("Channel: ") + msg.Channel + "\n"
		output += "\t\t" + utils.Green("Message: ") + msg.Content + "\n"
		output += "\t\t" + utils.Yellow("Attachments: ") + IntToString(len(msg.Attachments)) + "\n"
	}
	output += "\t" + utils.Muted("[/EXTERNAL MESSAGES]") + "\n"

	output += "\t" + utils.Muted("[RESPONSES]") + "\n"
	for _, resp := range cp.Responses {
		output += "\t\t" + utils.Blue("To: ") + resp.Recipient + "\n"
		output += "\t\t" + utils.Green("In Response To: ") + resp.OriginalMessageID + "\n"
		output += "\t\t" + utils.Yellow("Response: ") + resp.Content + "\n"
	}
	output += "\t" + utils.Muted("[/RESPONSES]") + "\n"

	output += "\t" + utils.Muted("[SUBSCRIPTIONS]") + "\n"
	for _, sub := range cp.Subscriptions {
		output += "\t\t" + utils.Blue("Topic: ") + sub.Topic + "\n"
		output += "\t\t" + utils.Green("Reason: ") + sub.Reason + "\n"
	}
	output += "\t" + utils.Muted("[/SUBSCRIPTIONS]") + "\n"

	output += "\t" + utils.Muted("[PENDING ACTIONS]") + "\n"
	for _, action := range cp.PendingActions {
		output += "\t\t" + utils.Blue("Action: ") + action.Description + "\n"
		output += "\t\t" + utils.Green("Priority: ") + action.Priority + "\n"
	}
	output += "\t" + utils.Muted("[/PENDING ACTIONS]") + "\n"

	output += "\t" + utils.Red("Done: ") + BoolToString(cp.Done) + "\n"
	output += utils.Dark("[/COMMUNICATION PLAN]") + "\n"
	return output
}

type Message struct {
	Recipient string `json:"recipient" jsonschema:"description=The recipient of the message"`
	Topic     string `json:"topic" jsonschema:"description=The topic of the message"`
	Content   string `json:"content" jsonschema:"description=The content of the message"`
}

type SlackMessage struct {
	Channel     string   `json:"channel" jsonschema:"description=The Slack channel to send the message to"`
	Content     string   `json:"content" jsonschema:"description=The content of the Slack message"`
	Attachments []string `json:"attachments" jsonschema:"description=Any attachments to include with the Slack message"`
}

type Response struct {
	Recipient         string `json:"recipient" jsonschema:"description=The recipient of the response"`
	OriginalMessageID string `json:"original_message_id" jsonschema:"description=The ID of the message being responded to"`
	Content           string `json:"content" jsonschema:"description=The content of the response"`
}

type Subscription struct {
	Topic  string `json:"topic" jsonschema:"description=The topic channel being subscribed to"`
	Reason string `json:"reason" jsonschema:"description=The reason for subscribing to this topic"`
}

type Action struct {
	Description string `json:"description" jsonschema:"description=A description of the communication action to be taken"`
	Priority    string `json:"priority" jsonschema:"description=The priority level of the action"`
}

func IntToString(i int) string {
	return fmt.Sprintf("%d", i)
}
