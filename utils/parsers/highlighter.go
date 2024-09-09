package parsers

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/ui"

	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/cue"
	"github.com/smacker/go-tree-sitter/dockerfile"
	"github.com/smacker/go-tree-sitter/elixir"
	"github.com/smacker/go-tree-sitter/elm"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/groovy"
	"github.com/smacker/go-tree-sitter/hcl"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/lua"
	tree_sitter_markdown "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown"
	"github.com/smacker/go-tree-sitter/ocaml"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/protobuf"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/sql"
	"github.com/smacker/go-tree-sitter/svelte"
	"github.com/smacker/go-tree-sitter/swift"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

/*
Highlighter uses tree-sitter to parse and highlight code for multiple languages.
We chose tree-sitter because it provides a fast and accurate way to parse code
and generate an Abstract Syntax Tree (AST), which is crucial for precise syntax highlighting.
This generic parser allows for easy extension to support multiple languages.
*/
type Highlighter struct {
	styles   *ui.Styles
	handle   *sitter.Parser
	language *sitter.Language
	code     string
	tree     *sitter.Tree
	ext      string
	err      error
}

/*
NewHighlighter creates a new Highlighter instance for the specified language.
We initialize the parser here but don't parse the code immediately,
allowing for more flexibility in when the parsing occurs.
*/
func NewHighlighter(code, ext string) *Highlighter {
	return &Highlighter{
		styles: ui.NewStyles(),
		handle: sitter.NewParser(),
		code:   code,
		ext:    ext,
	}
}

/*
ExtToLang returns the appropriate tree-sitter language based on the input string.
This function allows easy addition of new languages as they become supported.
*/
func (highlighter *Highlighter) ExtToLang(lang string) (*sitter.Language, error) {
	switch lang {
	case "go":
		return golang.GetLanguage(), nil
	case "js":
		return javascript.GetLanguage(), nil
	case "py":
		return python.GetLanguage(), nil
	case "c":
		return c.GetLanguage(), nil
	case "cpp":
		return cpp.GetLanguage(), nil
	case "sh":
		return bash.GetLanguage(), nil
	case "rb":
		return ruby.GetLanguage(), nil
	case "rs":
		return rust.GetLanguage(), nil
	case "java":
		return java.GetLanguage(), nil
	case "ts":
		return typescript.GetLanguage(), nil
	case "tsx":
		return tsx.GetLanguage(), nil
	case "php":
		return php.GetLanguage(), nil
	case "cs":
		return csharp.GetLanguage(), nil
	case "css":
		return css.GetLanguage(), nil
	case "dockerfile":
		return dockerfile.GetLanguage(), nil
	case "ex", "exs":
		return elixir.GetLanguage(), nil
	case "elm":
		return elm.GetLanguage(), nil
	case "hcl":
		return hcl.GetLanguage(), nil
	case "html":
		return html.GetLanguage(), nil
	case "kt":
		return kotlin.GetLanguage(), nil
	case "lua":
		return lua.GetLanguage(), nil
	case "ml", "mli":
		return ocaml.GetLanguage(), nil
	case "scala":
		return scala.GetLanguage(), nil
	case "sql":
		return sql.GetLanguage(), nil
	case "swift":
		return swift.GetLanguage(), nil
	case "toml":
		return toml.GetLanguage(), nil
	case "yaml", "yml":
		return yaml.GetLanguage(), nil
	case "cue":
		return cue.GetLanguage(), nil
	case "groovy":
		return groovy.GetLanguage(), nil
	case "md", "markdown":
		return tree_sitter_markdown.GetLanguage(), nil
	case "protobuf":
		return protobuf.GetLanguage(), nil
	case "svelte":
		return svelte.GetLanguage(), nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}

/*
Parse processes the code and generates the syntax tree.
We use a background context here as parsing is typically a quick operation,
but in more complex scenarios, this allows for potential timeout or cancellation.
*/
func (highlighter *Highlighter) Parse() *Highlighter {
	logger.Log("highlighter.Parse()")

	if highlighter.language, highlighter.err = highlighter.ExtToLang(highlighter.ext); highlighter.err != nil {
		logger.Error("highlighter.Parse() error: %s", highlighter.err.Error())
	}

	highlighter.handle.SetLanguage(highlighter.language)
	highlighter.tree, highlighter.err = highlighter.handle.ParseCtx(context.Background(), nil, []byte(highlighter.code))
	return highlighter
}

/*
Highlight traverses the syntax tree and applies appropriate styling to each node.
This method is the core of our syntax highlighting logic. It walks through the AST,
identifying different types of syntax elements and applying the corresponding styles.
A worker pool is used to process the nodes in parallel, which helps to improve performance,
but makes sure not to exhaust the available system resources.
The switch statement is used to map node types to specific styles. This approach
allows for easy extension if we need to add more specific highlighting rules in the future.
*/
func (highlighter *Highlighter) Highlight() string {
	rootNode := highlighter.tree.RootNode()
	var highlightedCode string

	type task struct {
		order int
		node  *sitter.Node
	}

	type result struct {
		order int
		text  string
	}

	numWorkers := 4 // Number of worker goroutines
	tasks := make(chan task, rootNode.NamedChildCount())
	results := make(chan result, rootNode.NamedChildCount())

	// Worker function
	worker := func() {
		for task := range tasks {
			text := highlighter.processNode(task.node)
			results <- result{order: task.order, text: text}
		}
	}

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	// Queue up tasks
	for i := 0; i < int(rootNode.NamedChildCount()); i++ {
		tasks <- task{order: i, node: rootNode.NamedChild(i)}
	}
	close(tasks)

	// Collect results
	orderedResults := make([]string, rootNode.NamedChildCount())
	for i := 0; i < int(rootNode.NamedChildCount()); i++ {
		res := <-results
		orderedResults[res.order] = res.text
	}
	close(results)

	// Concatenate ordered results
	for _, text := range orderedResults {
		highlightedCode += text
	}

	return highlightedCode
}

/*
processNode is the actual work that is done for each node in the tree.
It takes a node and returns a string with the appropriate style applied.
*/
func (highlighter *Highlighter) processNode(node *sitter.Node) string {
	text := node.Content([]byte(highlighter.code))
	var highlightedText string

	switch node.Type() {
	case "identifier", "type_declaration":
		highlightedText = highlighter.styles.VariableStyle.Render(text)
	case "number", "string", "char", "boolean":
		highlightedText = highlighter.styles.LiteralStyle.Render(text)
	case "keyword", "type":
		highlightedText = highlighter.styles.KeywordStyle.Render(text)
	case "comment":
		highlightedText = highlighter.styles.CommentStyle.Render(text)
	case "function", "method", "method_declaration", "function_declaration":
		highlightedText = highlighter.styles.FunctionStyle.Render(text)
	default:
		logger.Warn("unhandled node.Type(): %s", node.Type())
		highlightedText = text
	}

	for j := 0; j < int(node.NamedChildCount()); j++ {
		highlightedText += highlighter.processNode(node.NamedChild(j))
	}

	return highlightedText
}
