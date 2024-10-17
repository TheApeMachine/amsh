package mastercomputer

import (
	"context"
	"errors"
	"strings"

	"github.com/openai/openai-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

type Messaging struct {
	worker *Worker
}

func NewMessaging(worker *Worker) *Messaging {
	return &Messaging{worker: worker}
}

func (messaging *Messaging) ID() string {
	return messaging.worker.ID()
}

func (messaging *Messaging) Name() string {
	return "Messaging"
}

func (messaging *Messaging) Ctx() context.Context {
	return messaging.worker.Ctx()
}

func (messaging *Messaging) Manager() *twoface.WorkerManager {
	return messaging.worker.Manager()
}

func (messaging *Messaging) Reply(message *data.Artifact) {
	filters := messaging.worker.buffer.Peek("filters")

	// AI's should not act like schizos.
	if message.Peek("origin") == messaging.worker.name {
		return
	}

	if filters != "" && message.Peek("scope") != messaging.worker.buffer.Peek("origin") {
		filters := strings.Split(filters, ",")
		for _, filter := range filters {
			if filter == message.Peek("scope") {
				return
			}
		}
	}

	canAccept := messaging.worker.IsAllowed(WorkerStateAccepted)

	out := "ACKNOWLEDGED - "

	if canAccept {
		out += "ACCEPTED"
	} else {
		out += "REJECTED"
	}

	reply := viper.GetViper().GetString("messaging.templates.reply")
	reply = strings.ReplaceAll(reply, "{id}", messaging.ID())
	reply = strings.ReplaceAll(reply, "{sender}", messaging.worker.name)
	reply = strings.ReplaceAll(reply, "{message}", out)

	payload := message.Peek("payload")
	replyMsg := data.New(messaging.ID(), "reply", message.Peek("origin"), []byte(
		strings.Join([]string{payload, reply}, "\n\n"),
	))

	messaging.worker.queue.Publish(replyMsg)
	messaging.worker.state = WorkerStateAccepted
	messaging.worker.buffer.Poke("user", message.Peek("user"))
	messaging.worker.buffer.Poke("payload", message.Peek("payload"))
}

func (messaging *Messaging) Call(args map[string]any, owner twoface.Process) (string, error) {
	var (
		topic   string
		message string
		ok      bool
	)

	if topic, ok = args["topic"].(string); !ok {
		return "", errors.New("topic is not a string")
	}

	if message, ok = args["message"].(string); !ok {
		return "", errors.New("message is not a string")
	}

	artifact := data.New(owner.Name(), "message", topic, []byte{})
	artifact.Poke("id", utils.NewID())
	artifact.Poke("payload", message)
	twoface.NewQueue().Publish(artifact)
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

type SubscribeTopic struct {
	ctx context.Context
}

func NewSubscribeTopic(ctx context.Context) *SubscribeTopic {
	return &SubscribeTopic{ctx: ctx}
}

func (subscribe *SubscribeTopic) ID() string {
	return utils.NewID()
}

func (subscribe *SubscribeTopic) Name() string {
	return "SubscribeTopic"
}

func (subscribe *SubscribeTopic) Ctx() context.Context {
	return subscribe.ctx
}

func (subscribe *SubscribeTopic) Call(args map[string]any, owner twoface.Process) (string, error) {
	topic := args["topic"].(string)
	twoface.NewQueue().Subscribe(owner.Name(), topic)
	return "", nil
}

func (subscribe *SubscribeTopic) Schema() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"subscribe_topic",
		"Subscribe to a topic channel. You must be subscribed to the channel to publish to it.",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"topic": map[string]string{
					"type":        "string",
					"description": "The topic channel you want to subscribe to.",
				},
			},
			"required": []string{"topic"},
		},
	)
}
