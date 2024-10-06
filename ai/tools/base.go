package tools

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

/*
Tool is an interface for any objects that want to be used as tools.
*/
type Tool interface {
	Instructions() string
	Use(ctx context.Context, command string, params fiber.Map) (string, error)
}
