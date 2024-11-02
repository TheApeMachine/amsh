package parser

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"

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
	markdown "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown"
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
TreeSitterParser is used to parse many different coding languages, and
extract the syntax tree, so we can build up a graph of the code.
This provides a path for AI developers to get a better understanding of the code.
*/
type TreeSitterParser struct {
	handle      *sitter.Parser
	ignorePaths []string
	language    *sitter.Language
	err         error
	conn        *tools.Neo4j
}

/*
NewTreeSitterParser creates a new TreeSitterParser instance.
*/
func NewTreeSitterParser() *TreeSitterParser {
	return &TreeSitterParser{
		handle: sitter.NewParser(),
		conn:   tools.NewNeo4j(),
		ignorePaths: []string{
			"node_modules",
			"vendor",
			"dist",
			"build",
			"target",
			"tmp",
			"cache",
			"logs",
			"docs",
			"examples",
			".git",
			"lib",
			"switchery",
			"tinycon",
			"PHPExcel",
		},
	}
}

/*
ExtToLang converts a file extension to a TreeSitter language.
*/
func (parser *TreeSitterParser) ExtToLang(ext string) (*sitter.Language, error) {
	switch ext {
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
		return markdown.GetLanguage(), nil
	case "proto", "protobuf":
		return protobuf.GetLanguage(), nil
	case "svelte":
		return svelte.GetLanguage(), nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", ext)
	}
}

/*
Parse processes the code and generates the syntax tree.
We use a background context here as parsing is typically a quick operation,
but in more complex scenarios, this allows for potential timeout or cancellation.
*/
func (parser *TreeSitterParser) Parse(file *os.File) *TreeSitterParser {
	// Get the extension of the file
	ext := strings.TrimPrefix(filepath.Ext(file.Name()), ".")
	errnie.Info(ext)

	if parser.language, parser.err = parser.ExtToLang(ext); parser.err != nil {
		errnie.Error(parser.err)
	}

	buf := errnie.SafeMust(func() ([]byte, error) { return io.ReadAll(file) })

	parser.handle.SetLanguage(parser.language)
	tree := errnie.SafeMust(func() (*sitter.Tree, error) {
		return parser.handle.ParseCtx(context.Background(), nil, buf)
	})

	parser.processNode(tree.RootNode(), buf)

	return parser
}

/*
processNode handles the mapping of nodes and edges to turn the syntax tree into a graph.
*/
func (parser *TreeSitterParser) processNode(node *sitter.Node, buf []byte) {
	switch node.Type() {
	case "identifier", "type_declaration", "type_identifier":
		// Add a node to the graph.
		parser.conn.Use(context.Background(), map[string]any{
			"cypher": fmt.Sprintf(`CREATE (n:Node {
				name: "%s",
				content: "%s",
			})`, node.Content(buf), node.Content(buf)),
		})
	case "number", "string", "char", "boolean":
	case "keyword", "type":
	case "comment":
	case "function", "method", "method_declaration", "function_declaration":
		// Add a node to the graph.
		parser.conn.Use(context.Background(), map[string]any{
			"cypher": fmt.Sprintf(`CREATE (n:Node {
				name: "%s",
				content: "%s",
			})`, node.Content(buf), node.Content(buf)),
		})
	case "call_expression":
		parser.conn.Use(context.Background(), map[string]any{
			"cypher": fmt.Sprintf(`CREATE (n:Node {
				name: "%s",
				content: "%s",
			})`, node.Content(buf), node.Content(buf)),
		})

		// Add an edge to the graph.
		parser.conn.Use(context.Background(), map[string]any{
			"cypher": fmt.Sprintf(`CREATE (n:Edge {
				name: "%s",
				content: "%s",
			})`, node.Content(buf), node.Content(buf)),
		})
	case "return_statement":
		// Not sure what to do with this yet.
	default:
		errnie.Warn(node.Type())
	}

	for j := 0; j < int(node.NamedChildCount()); j++ {
		child := node.NamedChild(j)
		if child != nil {
			parser.processNode(child, buf)
		}
	}
}

/*
WalkDir walks through the directory and parses all the code file.
It ignores the paths in the ignorePaths list.
*/
func (parser *TreeSitterParser) WalkDir(dir string) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && slices.Contains(parser.ignorePaths, d.Name()) {
			return filepath.SkipDir
		}

		// Check if the file is a code file
		if !d.IsDir() && slices.Contains(codeFileExts, filepath.Ext(path)) {
			parser.Parse(errnie.SafeMust(func() (*os.File, error) { return os.Open(path) }))
		}

		return nil
	})
}

/*
codeFileExts is a list of file extensions that are considered code files.
*/
var codeFileExts = []string{
	".go",
	".js",
	".py",
	".c",
	".cpp",
	".sh",
	".rb",
	".rs",
	".java",
	".ts",
	".tsx",
	".php",
	".cs",
	".css",
	".html",
	".kt",
	".lua",
	".ml",
	".mli",
	".scala",
}
