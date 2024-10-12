package errnie

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/acarl005/stripansi"
	"github.com/davecgh/go-spew/spew"
	"github.com/openai/openai-go"
)

type ErrorHandler struct {
	fixing bool
	ai     *ErrorAI
	mu     sync.Mutex
}

type ErrorAI struct {
	client *openai.Client
}

func NewErrorAI() *ErrorAI {
	return &ErrorAI{
		client: openai.NewClient(),
	}
}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		fixing: false,
		ai:     NewErrorAI(),
	}
}

func (handler *ErrorHandler) Error(err error) error {
	if err == nil {
		return nil
	}

	handler.mu.Lock()
	defer handler.mu.Unlock()

	if handler.fixing {
		return fmt.Errorf("already fixing an issue")
	}

	handler.fixing = true
	defer func() { handler.fixing = false }()

	// Capture caller info
	file, line := captureCaller(2)
	message := fmt.Sprintf("❗ %s:%d %v", file, line, err)
	fmt.Println(message)
	writeToLog(message)

	codeSnippet := getCodeSnippet(file, line, 2)
	stackTrace := getStackTrace()
	fileContent := getFileContent(file)

	// Start the AI analysis loop
	ctx := context.Background()
	err = handler.ai.AnalyzeAndFix(ctx, err.Error(), codeSnippet, stackTrace, fileContent, file)
	if err != nil {
		fmt.Printf("Error during analysis and fix process: %v\n", err)
	}

	return fmt.Errorf(message)
}

func (ai *ErrorAI) AnalyzeAndFix(ctx context.Context, errMsg, snippet, stackTrace, fileContent, filePath string) error {
	var analysis ErrorAnalysis
	additionalContext := ""
	feedback := ""

	for {
		fmt.Println("Analyzing error...")
		response, err := ai.analyzeError(ctx, errMsg, snippet, stackTrace, fileContent, additionalContext, feedback)
		if err != nil {
			return fmt.Errorf("error analyzing: %v", err)
		}

		if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &analysis); err != nil {
			return fmt.Errorf("error parsing AI response: %v", err)
		}

		fmt.Println("Analysis Result:")
		for _, step := range analysis.Steps {
			fmt.Printf("Thought: %s\n", step.Thought)
			fmt.Printf("Missing: %s\n", step.Missing)
		}

		if len(analysis.Steps) > 0 && len(analysis.Steps[len(analysis.Steps)-1].Request.Filenames) > 0 || len(analysis.Steps[len(analysis.Steps)-1].Request.Searches) > 0 {
			additionalContext, err = ai.handleRequests([]Request{analysis.Steps[len(analysis.Steps)-1].Request})
			if err != nil {
				return fmt.Errorf("error handling requests: %v", err)
			}
			continue
		}

		if len(analysis.Fixes) == 0 {
			fmt.Println("No fixes suggested. Analysis complete.")
			return nil
		}

		updatedContent, newFeedback, err := ai.applyFixes(analysis.Fixes, fileContent, filePath)
		if err != nil {
			return fmt.Errorf("error applying fixes: %v", err)
		}

		if updatedContent == fileContent && newFeedback == "" {
			fmt.Println("No changes were applied and no feedback provided. Analysis complete.")
			return nil
		}

		fileContent = updatedContent
		feedback = newFeedback
		if feedback != "" {
			fmt.Println("Feedback recorded. Continuing analysis...")
		} else {
			fmt.Println("Fixes applied. Continuing analysis...")
		}
	}
}

func (ai *ErrorAI) analyzeError(ctx context.Context, errMsg, snippet, stackTrace, fileContent, additionalContext, feedback string) (*openai.ChatCompletion, error) {
	prompt := fmt.Sprintf("Error: %s\nSnippet: %s\nStackTrace: \n\n%s\n\nFileContent: \n\n%s\n\n", errMsg, snippet, stackTrace, fileContent)
	if additionalContext != "" {
		prompt += fmt.Sprintf("Additional Context: \n\n%s\n\n", additionalContext)
	}
	if feedback != "" {
		prompt += fmt.Sprintf("User Feedback: \n\n%s\n\n", feedback)
	}

	params := openai.ChatCompletionNewParams{
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.Float(0.0),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(`
			You are an expert at analyzing Go code and discovering the root cause of errors. 
			You should always try to be very sure of yourself by request each bit of context needed to verify any proposed fixes. 
			If you are not 100% sure, just say so, we can always look at it together.
			`),
			openai.UserMessage(prompt + `
			\n\n Let's think it through step by step.
			`),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F("error_analysis"),
					Description: openai.F("Error analysis and fixes"),
					Schema:      openai.F(GenerateSchema[ErrorAnalysis]()),
					Strict:      openai.Bool(false),
				}),
			},
		),
	}

	spew.Dump(params)
	os.Exit(1)

	return ai.client.Chat.Completions.New(ctx, params)
}

func (ai *ErrorAI) handleRequests(requests []Request) (string, error) {
	var additionalContext strings.Builder

	for _, request := range requests {
		if len(request.Filenames) > 0 {
			for _, filename := range request.Filenames {
				content := getFileContent(filename)
				additionalContext.WriteString(fmt.Sprintf("File: %s\nContent:\n%s\n\n", filename, content))
			}
		}

		if len(request.Searches) > 0 {
			for _, search := range request.Searches {
				results, err := searchCodebase(search)
				if err != nil {
					return "", fmt.Errorf("error searching codebase: %v", err)
				}
				additionalContext.WriteString(fmt.Sprintf("Search results for: %s\n%s\n\n", search, results))
			}
		}
	}

	return additionalContext.String(), nil
}

func (ai *ErrorAI) applyFixes(fixes []Fix, fileContent, filePath string) (string, string, error) {
	updatedContent := fileContent
	reader := bufio.NewReader(os.Stdin)
	var feedback strings.Builder

	for _, fix := range fixes {
		fmt.Printf("Suggested fix:\nReplace:\n%s\nWith:\n%s\nReason: %s\n", fix.Old, fix.New, fix.Why)
		fmt.Print("Apply this fix? (y/n/e): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		switch response {
		case "y":
			updatedContent = strings.Replace(updatedContent, fix.Old, fix.New, -1)
			fmt.Println("Fix applied.")
		case "n":
			fmt.Println("Fix skipped.")
			feedback.WriteString(fmt.Sprintf("Fix skipped: %s\n", fix.Why))
		case "e":
			fmt.Print("Enter your explanation or feedback: ")
			userFeedback, _ := reader.ReadString('\n')
			feedback.WriteString(fmt.Sprintf("User feedback for fix (%s): %s\n", fix.Why, userFeedback))
			fmt.Println("Feedback recorded. Continuing with next fix.")
		default:
			fmt.Println("Invalid response. Skipping this fix.")
			feedback.WriteString(fmt.Sprintf("Fix skipped (invalid response): %s\n", fix.Why))
		}
	}

	if updatedContent != fileContent {
		if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
			return "", "", fmt.Errorf("error writing updated content to file: %v", err)
		}
		fmt.Println("File updated successfully.")
	}

	return updatedContent, feedback.String(), nil
}

func captureCaller(skip int) (file string, line int) {
	var pc [10]uintptr
	n := runtime.Callers(skip, pc[:])
	if n == 0 {
		return "", 0
	}

	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.File, frame.Line
}

func getFileContent(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(content)
}

func getCodeSnippet(file string, line, radius int) string {
	fileHandle, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer fileHandle.Close()

	scanner := bufio.NewScanner(fileHandle)
	currentLine := 1
	var snippet string

	for scanner.Scan() {
		if currentLine >= line-radius && currentLine <= line+radius {
			prefix := "  "
			if currentLine == line {
				prefix = "> "
			}
			snippet += fmt.Sprintf("%s%d: %s\n", prefix, currentLine, scanner.Text())
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return ""
	}

	return snippet
}

func getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func writeToLog(message string) {
	logFileMu.Lock()
	defer logFileMu.Unlock()

	if logFile != nil {
		fmt.Fprintln(logFile, stripansi.Strip(message))
	}
}

func searchCodebase(query string) (string, error) {
	cmd := exec.Command("ack", "-n", "--color", query)
	cmd.Dir = "." // Set the working directory to the current directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// ack returns exit code 1 if no matches are found, which is not an error for us
			if exitError.ExitCode() == 1 {
				return "No matches found.", nil
			}
		}
		return "", fmt.Errorf("error executing ack: %v", err)
	}

	// Limit the output to a reasonable size (e.g., first 1000 characters)
	const maxOutputSize = 1000
	result := string(output)
	if len(result) > maxOutputSize {
		result = result[:maxOutputSize] + "...\n(output truncated)"
	}

	return result, nil
}
