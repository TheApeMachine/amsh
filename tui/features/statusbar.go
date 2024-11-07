package features

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"
	"github.com/theapemachine/amsh/tui/types"
)

type StatusBar struct {
	model statusbar.Model
}

func NewStatusBar() *StatusBar {
	sb := &StatusBar{
		model: statusbar.New(
			statusbar.ColorConfig{
				Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#ffffff"},
				Background: lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F25D94"},
			},
			statusbar.ColorConfig{
				Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
				Background: lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#3c3836"},
			},
			statusbar.ColorConfig{
				Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
				Background: lipgloss.AdaptiveColor{Light: "#A550DF", Dark: "#A550DF"},
			},
			statusbar.ColorConfig{
				Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
				Background: lipgloss.AdaptiveColor{Light: "#6124DF", Dark: "#6124DF"},
			},
		),
	}

	// Initialize with default content
	sb.model.SetContent(
		"NORMAL",  // mode
		"No File", // filename
		"~",       // directory
		"1:1",     // position
	)

	return sb
}

func (sb *StatusBar) Model() tea.Model {
	return sb
}

func (sb *StatusBar) Name() string {
	return "statusbar"
}

func (sb *StatusBar) Size() (int, int) {
	return sb.model.Width, sb.model.Height
}

func (sb *StatusBar) Init() tea.Cmd {
	return nil
}

// Add new message type for mode changes
type ModeChangeMsg struct {
	Mode types.Mode
}

func (sb *StatusBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sb.model.SetSize(msg.Width)
		return sb, nil

	case FileSelectedMsg:
		// Update status bar with selected file
		sb.model.SetContent(
			"NORMAL", // mode
			msg.Path, // filename
			"~",      // directory
			"1:1",    // position
		)
		return sb, nil

	case LoadFileMsg:
		// Handle LoadFileMsg to keep status bar in sync
		sb.model.SetContent(
			"NORMAL",
			msg.Filepath,
			"~",
			"1:1",
		)
		return sb, nil

	case ModeChangeMsg:
		// Get current content
		filename := sb.model.SecondColumn
		directory := sb.model.ThirdColumn
		position := sb.model.FourthColumn

		// Update the mode display
		modeStr := "NORMAL"
		sb.model.FirstColumnColors.Background = lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#3c3836"}
		switch msg.Mode {
		case types.ModeInsert:
			sb.model.FirstColumnColors.Background = lipgloss.AdaptiveColor{Light: "#A550DF", Dark: "#A550DF"}
			modeStr = "INSERT"
		case types.ModeVisual:
			sb.model.FirstColumnColors.Background = lipgloss.AdaptiveColor{Light: "#6124DF", Dark: "#6124DF"}
			modeStr = "VISUAL"
		}

		// Update status bar preserving existing content
		sb.model.SetContent(
			modeStr,
			filename,
			directory,
			position,
		)
		return sb, nil
	}

	return sb, nil
}

func (sb *StatusBar) View() string {
	return sb.model.View()
}

func (sb *StatusBar) SetContent(width, height int) {
	sb.model.SetSize(width)
}
