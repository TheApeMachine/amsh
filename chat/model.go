package chat

import (
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/ui"
)

type Model struct {
	viewport viewport.Model
	textarea textarea.Model
	styles   *ui.Styles
	state    components.State
	buf      strings.Builder
	aiIO     *ai.IO
	reader   io.Reader
	writer   io.Writer
	width    int
	height   int
	err      error
}

func New(width, height int) *Model {
	ta := textarea.New()
	ta.SetHeight(height / 8)
	ta.SetWidth(width / 2)
	ta.ShowLineNumbers = false
	ta.Focus()

	pr, pw := io.Pipe()

	return &Model{
		viewport: viewport.New(width/2, height/4),
		textarea: ta,
		styles:   ui.NewStyles(),
		state:    components.Inactive,
		buf:      strings.Builder{},
		aiIO:     ai.NewIO(pr, pw),
		reader:   pr,
		writer:   pw,
		width:    width,
		height:   height,
	}
}

func (model *Model) Init() tea.Cmd {
	model.buf.WriteString("Type your message...")
	model.viewport.SetContent(model.buf.String())
	return model.viewport.Init()
}
