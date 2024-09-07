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
func (highlighter *Highlighter) Parse() {
	highlighter.handle.SetLanguage(highlighter.language)
	highlighter.tree, highlighter.err = highlighter.handle.ParseCtx(context.Background(), nil, []byte(highlighter.code))
}

/*
Highlight traverses the syntax tree and applies appropriate styling to each node.
This method is the core of our syntax highlighting logic. It walks through the AST,
identifying different types of syntax elements and applying the corresponding styles.

The switch statement is used to map node types to specific styles. This approach
allows for easy extension if we need to add more specific highlighting rules in the future.
*/
func (highlighter *Highlighter) Highlight() string {
	logger.Log("highlighter.Highlight()")

	if highlighter.err != nil {
		logger.Log("highlighter.Highlight() error: %s", highlighter.err.Error())
		return highlighter.code
	}

	language, err := highlighter.ExtToLang(highlighter.ext)
	if err != nil {
		logger.Log("highlighter.Highlight() error: %s", highlighter.err.Error())
		return highlighter.code
	}

	logger.Log("highlighter.Highlight() language: %s", highlighter.ext)

	highlighter.handle.SetLanguage(language)
	highlighter.tree, highlighter.err = highlighter.handle.ParseCtx(context.Background(), nil, []byte(highlighter.code))

	if highlighter.err != nil {
		logger.Log("highlighter.Highlight() error: %s", highlighter.err.Error())
		return highlighter.code
	}

	rootNode := highlighter.tree.RootNode()
	var highlightedCode string

	// We iterate through named children to focus on significant nodes,
	// skipping less important syntax elements like whitespace.
	for i := 0; i < int(rootNode.NamedChildCount()); i++ {
		child := rootNode.NamedChild(i)
		text := child.Content([]byte(highlighter.code))

		// Apply different styles based on the node type.
		// This switch could be expanded to handle more specific syntax elements.
		switch child.Type() {
		case "identifier":
			highlightedCode += ui.VariableStyle.Render(text)
		case "number", "string", "char", "boolean":
			highlightedCode += ui.LiteralStyle.Render(text)
		case "keyword", "type":
			highlightedCode += ui.KeywordStyle.Render(text)
		case "comment":
			highlightedCode += ui.CommentStyle.Render(text)
		case "function", "method":
			highlightedCode += ui.FunctionStyle.Render(text)
		default:
			// For unhandled types, we don't apply any styling.
			highlightedCode += text
		}
	}

	logger.Log("highlighter.Highlight() highlightedCode: %s", highlightedCode)
	return highlightedCode
}
