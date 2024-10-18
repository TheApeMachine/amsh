package mastercomputer

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
)

type Messaging struct {
	worker *Worker
}

func NewMessaging(worker *Worker) *Messaging {
	return &Messaging{worker: worker}
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
	reply = strings.ReplaceAll(reply, "{id}", message.Peek("id"))
	reply = strings.ReplaceAll(reply, "{sender}", messaging.worker.name)
	reply = strings.ReplaceAll(reply, "{message}", out)

	payload := message.Peek("payload")
	replyMsg := data.New(messaging.worker.name, "reply", message.Peek("origin"), []byte(
		strings.Join([]string{payload, reply}, "\n\n"),
	))

	messaging.worker.queue.Publish(replyMsg)
	messaging.worker.state = WorkerStateAccepted
	messaging.worker.buffer.Poke("payload", message.Peek("payload"))
}
