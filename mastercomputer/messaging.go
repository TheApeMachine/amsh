package mastercomputer

import (
	"strings"

	"github.com/openai/openai-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
)

type Messaging struct {
	worker *Worker
}

func NewMessaging(worker *Worker) *Messaging {
	return &Messaging{worker: worker}
}

func (messaging *Messaging) Reply(message *data.Artifact) {
	canAccept := messaging.worker.IsAllowed(WorkerStateAccepted)

	out := "ACKNOWLEDGED - "

	if canAccept {
		out += "ACCEPTED"
	} else {
		out += "REJECTED"
	}

	reply := viper.GetViper().GetString("messaging.templates.reply")
	reply = strings.ReplaceAll(reply, "{id}", messaging.worker.ID)
	reply = strings.ReplaceAll(reply, "{sender}", messaging.worker.name)
	reply = strings.ReplaceAll(reply, "{message}", out)

	payload := message.Peek("payload")
	replyMsg := data.New(messaging.worker.ID, "reply", message.Peek("origin"), []byte(
		strings.Join([]string{payload, reply}, "\n\n"),
	))

	messaging.worker.queue.Publish(replyMsg)
	messaging.worker.state = WorkerStateAccepted
}

func (messaging *Messaging) Call(args map[string]any) (string, error) {
	return "", nil
}

func (messaging *Messaging) Schema() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"publish_message",
		"Publish a message to a topic channel. You must be subscribed to the channel to publish to it.",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"topic": map[string]string{
					"type":        "string",
					"description": "The topic channel you want to post to.",
				},
				"message": map[string]string{
					"type":        "string",
					"description": "The content of the message you want to post.",
				},
			},
			"required": []string{"topic", "message"},
		},
	)
}
