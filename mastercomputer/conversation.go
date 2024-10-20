package mastercomputer

import (
	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
	"github.com/pkoukk/tiktoken-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
)

type Conversation struct {
	task             *data.Artifact
	context          []openai.ChatCompletionMessageParamUnion
	maxContextTokens int
	tokenCounts      []int64
}

func NewConversation(task *data.Artifact) *Conversation {
	return &Conversation{
		task: task,
		context: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(task.Peek("system")),
			openai.UserMessage(task.Peek("user") + "\n\n" + task.Peek("payload")),
		},
		maxContextTokens: viper.GetViper().GetInt("ai.max_context_tokens"),
		tokenCounts:      make([]int64, 0),
	}
}

func (conversation *Conversation) Update(message openai.ChatCompletionMessageParamUnion) {
	conversation.context = append(conversation.context, message)
}

func (conversation *Conversation) Truncate() []openai.ChatCompletionMessageParamUnion {
	maxTokens := conversation.maxContextTokens - 1024 // Reserve tokens for response
	totalTokens := 0
	var truncatedMessages []openai.ChatCompletionMessageParamUnion

	// Start from the most recent message
	for i := len(conversation.context) - 1; i >= 0; i-- {
		msg := conversation.context[i]
		messageTokens := conversation.estimateTokens(msg)
		if totalTokens+messageTokens <= maxTokens {
			truncatedMessages = append([]openai.ChatCompletionMessageParamUnion{msg}, truncatedMessages...)
			totalTokens += messageTokens
		} else {
			break
		}
	}

	return truncatedMessages
}

func (conversation *Conversation) UpdateTokenCounts(usage openai.CompletionUsage) {
	conversation.tokenCounts = append(conversation.tokenCounts, usage.TotalTokens)
}

func (conversation *Conversation) estimateTokens(msg openai.ChatCompletionMessageParamUnion) int {
	content := ""
	role := ""
	switch m := msg.(type) {
	case openai.ChatCompletionUserMessageParam:
		content = m.Content.String()
		role = "user"
	case openai.ChatCompletionAssistantMessageParam:
		content = m.Content.String()
		role = "assistant"
	case openai.ChatCompletionSystemMessageParam:
		content = m.Content.String()
		role = "system"
	case openai.ChatCompletionToolMessageParam:
		content = m.Content.String()
		role = "function"
	}

	// Use tiktoken-go to estimate tokens
	encoding, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		log.Error("Error getting encoding", "error", err)
		return 0
	}

	tokensPerMessage := 4 // As per OpenAI's token estimation guidelines

	numTokens := tokensPerMessage
	numTokens += len(encoding.Encode(content, nil, nil))
	if role == "user" || role == "assistant" || role == "system" || role == "function" {
		numTokens += len(encoding.Encode(role, nil, nil))
	}

	return numTokens
}
