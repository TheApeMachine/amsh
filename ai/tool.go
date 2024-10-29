package ai

type Tool interface {
	Use(arguments map[string]any) string
}

type ToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}
