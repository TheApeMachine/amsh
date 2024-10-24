package mastercomputer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
	"github.com/pkoukk/tiktoken-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/utils"
)

type Conversation struct {
	context          []openai.ChatCompletionMessageParamUnion
	maxContextTokens int
	tokenCounts      []int64
}

func NewConversation() *Conversation {
	return &Conversation{
		context:          []openai.ChatCompletionMessageParamUnion{},
		maxContextTokens: viper.GetViper().GetInt("ai.max_context_tokens"),
		tokenCounts:      make([]int64, 0),
	}
}

func (conversation *Conversation) Update(message openai.ChatCompletionMessageParamUnion) {
	// Format and print the message in a human-readable way
	printFormattedMessage(message)
	conversation.context = append(conversation.context, message)
}

// Helper function to format and print messages
func printFormattedMessage(msg openai.ChatCompletionMessageParamUnion) {
	var role, content string

	// Extract role and content based on message type
	switch m := msg.(type) {
	case openai.ChatCompletionSystemMessageParam:
		role = "System"
		content = m.Content.String()
	case openai.ChatCompletionUserMessageParam:
		role = "User"
		content = m.Content.String()
	case openai.ChatCompletionAssistantMessageParam:
		role = "Assistant"
		content = m.Content.String()
	case openai.ChatCompletionToolMessageParam:
		role = "Tool"
		content = m.Content.String()
	default:
		return
	}

	// Print formatted message with role as header
	fmt.Printf(
		"\n%s %s\n",
		utils.Muted(fmt.Sprintf("┌─── %s Message ───────────────────────────\n", role)),
		utils.Highlight(formatContent(content)),
	)
	fmt.Println(utils.Muted("└────────────────────────────────────────────\n"))
}

// Helper function to format content with proper line breaks and indentation
func formatContent(content string) string {
	// Replace newlines with newline + pipe + space for consistent formatting
	formatted := ""
	for i, line := range strings.Split(content, "\n") {
		if i > 0 {
			formatted += "\n│ "
		}
		formatted += line
	}
	return formatted
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
	case openai.ChatCompletionSystemMessageParam:
		content = m.Content.String()
		role = "system"
	case openai.ChatCompletionUserMessageParam:
		content = m.Content.String()
		role = "user"
	case openai.ChatCompletionAssistantMessageParam:
		content = m.Content.String()
		role = "assistant"
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
