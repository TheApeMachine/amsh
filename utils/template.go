package utils

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
)

func ComposedMessage(id string, buffer data.Artifact) (system string, user string) {
	v := viper.GetViper()
	system = strings.ReplaceAll(v.GetString("ai.prompt.system"), "{id}", id)
	system = strings.ReplaceAll(system, "{role}", v.GetString("ai.prompt.role"))
	system = strings.ReplaceAll(system, "{guidelines}", v.GetString("ai.prompt.guidelines"))

	user = viper.GetViper().GetString("ai.prompt." + buffer.Peek("format"))
	user = strings.ReplaceAll(user, "{id}", id)

	return
}

func MessageTemplate(id, sender, topic, message string) string {
	tmpl := viper.GetViper().GetString("messaging.templates.new")
	tmpl = strings.ReplaceAll(tmpl, "{id}", id)
	tmpl = strings.ReplaceAll(tmpl, "{sender}", sender)
	tmpl = strings.ReplaceAll(tmpl, "{topic}", topic)
	tmpl = strings.ReplaceAll(tmpl, "{message}", message)

	return tmpl
}
