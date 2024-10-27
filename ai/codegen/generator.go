package codegen

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/theapemachine/amsh/ai/types"
	"github.com/theapemachine/amsh/container"
)

type Generator struct {
	builder *container.Builder
	runner  *container.Runner
}

func NewGenerator() (*Generator, error) {
	builder, err := container.NewBuilder()
	if err != nil {
		return nil, fmt.Errorf("failed to create builder: %w", err)
	}

	runner, err := container.NewRunner()
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &Generator{
		builder: builder,
		runner:  runner,
	}, nil
}

type CodeRequest struct {
	Language    string            `json:"language"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Imports     []string          `json:"imports"`
	Interfaces  []InterfaceSpec   `json:"interfaces,omitempty"`
	Structs     []StructSpec      `json:"structs,omitempty"`
	Functions   []FunctionSpec    `json:"functions,omitempty"`
	Tests       bool              `json:"tests"`
	Config      map[string]string `json:"config,omitempty"`
}

type InterfaceSpec struct {
	Name    string       `json:"name"`
	Methods []MethodSpec `json:"methods"`
	Doc     []string     `json:"doc,omitempty"`
	Embed   []string     `json:"embed,omitempty"`
}

type StructSpec struct {
	Name   string      `json:"name"`
	Fields []FieldSpec `json:"fields"`
	Doc    []string    `json:"doc,omitempty"`
	Embed  []string    `json:"embed,omitempty"`
}

type MethodSpec struct {
	Name       string      `json:"name"`
	Params     []FieldSpec `json:"params"`
	Returns    []FieldSpec `json:"returns"`
	Doc        []string    `json:"doc,omitempty"`
	Visibility string      `json:"visibility"`
}

type FunctionSpec struct {
	Name       string      `json:"name"`
	Params     []FieldSpec `json:"params"`
	Returns    []FieldSpec `json:"returns"`
	Doc        []string    `json:"doc,omitempty"`
	Visibility string      `json:"visibility"`
	Body       string      `json:"body,omitempty"`
}

type FieldSpec struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Tag        string `json:"tag,omitempty"`
	Doc        string `json:"doc,omitempty"`
	Visibility string `json:"visibility"`
}

type GenerationResult struct {
	Code     string      `json:"code"`
	FilePath string      `json:"file_path"`
	PR       PullRequest `json:"pull_request"`
}

type PullRequest struct {
	ID          string   `json:"id"`
	Branch      string   `json:"branch"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Reviews     []Review `json:"reviews"`
}

type Review struct {
	ID        string    `json:"id"`
	Reviewer  string    `json:"reviewer"`
	Comments  []Comment `json:"comments"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

func (g *Generator) Generate(ctx context.Context, req CodeRequest) (*GenerationResult, error) {
	// Build the development container with necessary tools
	imageName := "codegen:latest"
	if err := g.builder.BuildImage(ctx, "./container/codegen", imageName); err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}

	// Start the container
	cmd := []string{"/bin/bash"}
	username := "codegen"
	customMsg := "Code generation environment ready"

	in, out, err := g.runner.RunContainer(ctx, imageName, cmd, username, customMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	defer in.Close()
	defer out.Close()

	// Generate the code
	code, err := g.generateCode(req)
	if err != nil {
		return nil, err
	}

	// Create a new branch
	branchName := fmt.Sprintf("aigenerated/%s_%s", req.Name, time.Now().Format("20060102150405"))
	if output := g.runner.ExecuteCommand(ctx, []string{
		"git", "checkout", "-b", branchName,
	}); output == nil {
		return nil, fmt.Errorf("failed to create branch")
	}

	// Write the generated code to file
	filePath := fmt.Sprintf("%s/%s.go", req.Name, req.Name)
	if output := g.runner.ExecuteCommand(ctx, []string{
		"bash", "-c", fmt.Sprintf("echo '%s' > %s", code, filePath),
	}); output == nil {
		return nil, fmt.Errorf("failed to write code")
	}

	// Create PR
	pr, err := g.createPullRequest(ctx, branchName, req, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return &GenerationResult{
		Code:     code,
		FilePath: filePath,
		PR:       *pr,
	}, nil
}

func (g *Generator) generateCode(req CodeRequest) (string, error) {
	var code strings.Builder

	// Add package declaration
	code.WriteString(fmt.Sprintf("package %s\n\n", req.Name))

	// Add imports
	if len(req.Imports) > 0 {
		code.WriteString("import (\n")
		for _, imp := range req.Imports {
			code.WriteString(fmt.Sprintf("\t%q\n", imp))
		}
		code.WriteString(")\n\n")
	}

	// Add interfaces
	for _, iface := range req.Interfaces {
		// Add documentation
		for _, doc := range iface.Doc {
			code.WriteString(fmt.Sprintf("// %s\n", doc))
		}
		code.WriteString(fmt.Sprintf("type %s interface {\n", iface.Name))
		for _, method := range iface.Methods {
			code.WriteString(fmt.Sprintf("\t%s(", method.Name))
			// Add parameters
			for i, param := range method.Params {
				if i > 0 {
					code.WriteString(", ")
				}
				code.WriteString(fmt.Sprintf("%s %s", param.Name, param.Type))
			}
			code.WriteString(")")
			// Add returns
			if len(method.Returns) > 0 {
				if len(method.Returns) == 1 && method.Returns[0].Name == "" {
					code.WriteString(fmt.Sprintf(" %s", method.Returns[0].Type))
				} else {
					code.WriteString(" (")
					for i, ret := range method.Returns {
						if i > 0 {
							code.WriteString(", ")
						}
						if ret.Name != "" {
							code.WriteString(fmt.Sprintf("%s ", ret.Name))
						}
						code.WriteString(ret.Type)
					}
					code.WriteString(")")
				}
			}
			code.WriteString("\n")
		}
		code.WriteString("}\n\n")
	}

	// Add structs
	for _, str := range req.Structs {
		// Add documentation
		for _, doc := range str.Doc {
			code.WriteString(fmt.Sprintf("// %s\n", doc))
		}
		code.WriteString(fmt.Sprintf("type %s struct {\n", str.Name))
		for _, field := range str.Fields {
			if field.Tag != "" {
				code.WriteString(fmt.Sprintf("\t%s %s `%s`\n", field.Name, field.Type, field.Tag))
			} else {
				code.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, field.Type))
			}
		}
		code.WriteString("}\n\n")
	}

	// Add functions
	for _, fn := range req.Functions {
		// Add documentation
		for _, doc := range fn.Doc {
			code.WriteString(fmt.Sprintf("// %s\n", doc))
		}
		code.WriteString(fmt.Sprintf("func %s(", fn.Name))
		// Add parameters
		for i, param := range fn.Params {
			if i > 0 {
				code.WriteString(", ")
			}
			code.WriteString(fmt.Sprintf("%s %s", param.Name, param.Type))
		}
		code.WriteString(")")
		// Add returns
		if len(fn.Returns) > 0 {
			if len(fn.Returns) == 1 && fn.Returns[0].Name == "" {
				code.WriteString(fmt.Sprintf(" %s", fn.Returns[0].Type))
			} else {
				code.WriteString(" (")
				for i, ret := range fn.Returns {
					if i > 0 {
						code.WriteString(", ")
					}
					if ret.Name != "" {
						code.WriteString(fmt.Sprintf("%s ", ret.Name))
					}
					code.WriteString(ret.Type)
				}
				code.WriteString(")")
			}
		}
		code.WriteString(" {\n")
		if fn.Body != "" {
			code.WriteString(fn.Body)
		}
		code.WriteString("}\n\n")
	}

	return code.String(), nil
}

func (g *Generator) createPullRequest(ctx context.Context, branch string, req CodeRequest, filePath string) (*PullRequest, error) {
	// Stage and commit the changes
	cmds := [][]string{
		{"git", "add", filePath},
		{"git", "commit", "-m", fmt.Sprintf("AI: Generate %s\n\n%s", req.Name, req.Description)},
		{"git", "push", "origin", branch},
	}

	for _, cmd := range cmds {
		if output := g.runner.ExecuteCommand(ctx, cmd); output == nil {
			return nil, fmt.Errorf("git command failed: %s", strings.Join(cmd, " "))
		}
	}

	// Create PR using GitHub API
	prCmd := []string{
		"gh", "pr", "create",
		"--title", fmt.Sprintf("AI: Generate %s", req.Name),
		"--body", fmt.Sprintf("Generated code for %s\n\n%s", req.Name, req.Description),
		"--base", "main",
		"--head", branch,
	}

	if output := g.runner.ExecuteCommand(ctx, prCmd); output == nil {
		return nil, fmt.Errorf("failed to create PR")
	}

	return &PullRequest{
		ID:          fmt.Sprintf("pr_%d", time.Now().UnixNano()),
		Branch:      branch,
		Title:       fmt.Sprintf("AI: Generate %s", req.Name),
		Description: req.Description,
		Status:      "pending_review",
		Reviews:     []Review{},
	}, nil
}

// Execute implements the Tool interface
func (g *Generator) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Convert args to CodeRequest
	req := CodeRequest{
		Language:    getStringArg(args, "language", "Go"),
		Name:        getStringArg(args, "name", "generated"),
		Description: getStringArg(args, "description", ""),
		Tests:       getBoolArg(args, "tests", false),
	}

	// Generate the code using the request
	result, err := g.Generate(ctx, req)
	if err != nil {
		return "", err
	}

	return result.Code, nil
}

// GetSchema implements the Tool interface
func (g *Generator) GetSchema() types.ToolSchema {
	return types.ToolSchema{
		Name:        "code_generator",
		Description: "Generates code based on provided specifications",
		Parameters: map[string]interface{}{
			"language": map[string]interface{}{
				"type":        "string",
				"description": "Programming language to generate code in",
				"default":     "Go",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the code artifact to generate",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description of what the code should do",
			},
			"tests": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to generate tests",
				"default":     false,
			},
		},
	}
}

// Helper functions for arg extraction
func getStringArg(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultValue
}

func getBoolArg(args map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := args[key].(bool); ok {
		return val
	}
	return defaultValue
}
