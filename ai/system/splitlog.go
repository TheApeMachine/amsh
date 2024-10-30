package system

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
const useHighPerformanceRenderer = false

type SplitLog struct {
	ready  bool
	pm     *ProcessManager
	prompt string
	agents map[string][]string
	index  int
}

func NewSplitLog(pm *ProcessManager, prompt string) *SplitLog {
	return &SplitLog{
		pm:     pm,
		prompt: prompt,
		agents: map[string][]string{},
		index:  0,
	}
}

func (model *SplitLog) Init() tea.Cmd {
	go func() {
		for event := range model.pm.Execute(model.prompt) {
			// Check if the agents map has the agent name as a key.
			if _, ok := model.agents[event.AgentID]; !ok {
				model.agents[event.AgentID] = []string{}
			}
			model.agents[event.AgentID] = append(model.agents[event.AgentID], event.Content)
		}
	}()
	return nil
}

func (model *SplitLog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return model, tea.Quit
		}

	case tea.WindowSizeMsg:
		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			// cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	cmds = append(cmds, cmd)

	return model, tea.Batch(cmds...)
}

func (model *SplitLog) View() string {
	if !model.ready {
		return "\n  Initializing..."
	}

	columns := []string{}
	for _, agent := range model.agents {
		columns = append(columns, strings.Join(agent, "\n"))
	}
	return lipgloss.JoinVertical(lipgloss.Left, columns...)
}
