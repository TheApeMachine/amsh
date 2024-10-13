package tools

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/integration/boards"
)

/*
Project is a wrapper to make tools from various project related systems.
It allows agents to search for information about projects, tasks, and other related information.
*/
type Project struct {
}

/*
NewProject creates a new project tool.
*/
func NewProject() *Project {
	return &Project{}
}

/*
Instructions to be used in the system or user prompt for an Agent.
*/
func (project *Project) Instructions() string {
	return viper.GetViper().GetString("tools.project.instructions")
}

/*
User the project tool
*/
func (project *Project) Use(
	ctx context.Context, command string, params fiber.Map,
) (string, error) {
	switch command {
	case "search":
		boardsSrv := boards.NewService()
		boardsSrv.SearchWorkitems(ctx, params["query"].(string))
	default:
		return "the command you provided does not exist", errors.New("invalid command")
	}

	return "", nil
}
