package tools

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/slack-go/slack"
	"github.com/theapemachine/amsh/ai/types"
)

type SlackTool struct {
	api *slack.Client
}

func NewSlackTool() *SlackTool {
	botToken := os.Getenv("BOT_TOKEN")
	return &SlackTool{
		api: slack.New(botToken),
	}
}

func (s *SlackTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	operation, err := getStringArg(args, "operation", "")
	if err != nil || operation == "" {
		return "", errors.New("operation is required")
	}

	switch operation {
	case "send":
		return s.sendMessage(ctx, args)
	case "search":
		return s.searchMessages(ctx, args)
	default:
		return "", fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (s *SlackTool) sendMessage(ctx context.Context, args map[string]interface{}) (string, error) {
	channelID, err := getStringArg(args, "id", "")
	if err != nil || channelID == "" {
		return "", errors.New("channel ID is required for send operation")
	}

	message, err := getStringArg(args, "message", "")
	if err != nil || message == "" {
		return "", errors.New("message is required for send operation")
	}

	_, _, err = s.api.PostMessageContext(ctx, channelID, slack.MsgOptionText(message, false))
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return "Message sent successfully", nil
}

func (s *SlackTool) searchMessages(ctx context.Context, args map[string]interface{}) (string, error) {
	query, err := getStringArg(args, "keywords", "")
	if err != nil || query == "" {
		return "", errors.New("keywords are required for search operation")
	}

	searchParams := slack.NewSearchParameters()
	results, err := s.api.SearchMessagesContext(ctx, query, searchParams)
	if err != nil {
		return "", fmt.Errorf("failed to search messages: %w", err)
	}

	if len(results.Matches) == 0 {
		return "No messages found for the given query.", nil
	}

	output := "Search Results:\n"
	for _, msg := range results.Matches {
		output += fmt.Sprintf("- Channel: %s\n  User: %s\n  Text: %s\n\n", msg.Channel.Name, msg.Username, msg.Text)
	}

	return output, nil
}

func (s *SlackTool) GetSchema() types.ToolSchema {
	return types.ToolSchema{
		Name:        "slack",
		Description: "Interact with Slack.",
		Parameters: map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The operation to perform on the Slack messages or channels.",
				"enum":        []string{"search", "send"},
			},
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the Slack channel or user to send a message to.",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The message to send to the Slack channel or user.",
			},
			"keywords": map[string]interface{}{
				"type":        "string",
				"description": "The keywords or valid Slack search query to search for messages.",
			},
		},
	}
}
