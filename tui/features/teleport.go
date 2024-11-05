package features

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Target struct {
	Line   int
	Col    int
	Code   string
	Weight int
}

type Teleport struct {
	active  bool
	targets []Target
	input   string
}

// TeleportMsg is sent when a teleport target is selected
type TeleportMsg struct {
	Line, Col int
}

func NewTeleport() *Teleport {
	return &Teleport{
		active:  false,
		targets: make([]Target, 0),
	}
}

// Toggle activates or deactivates teleport mode and returns a command
func (teleport *Teleport) Toggle() tea.Cmd {
	teleport.active = !teleport.active
	teleport.input = ""
	if teleport.active {
		return teleport.Analyze
	}
	return nil
}

// IsActive returns whether teleport mode is active
func (teleport *Teleport) IsActive() bool {
	return teleport.active
}

// AddInput adds a character to the input sequence and returns a command if matched
func (teleport *Teleport) AddInput(char rune) tea.Cmd {
	teleport.input += string(char)

	// Check for exact match
	for _, t := range teleport.targets {
		if t.Code == teleport.input {
			return func() tea.Msg {
				return TeleportMsg{Line: t.Line, Col: t.Col}
			}
		}
	}

	// Check if any targets still match the partial input
	matching := false
	for _, t := range teleport.targets {
		if len(t.Code) >= len(teleport.input) && t.Code[:len(teleport.input)] == teleport.input {
			matching = true
			break
		}
	}

	// Reset input if no matches
	if !matching {
		teleport.input = ""
	}

	return nil
}

// Analyze is now a tea.Cmd that analyzes the textarea content
func (teleport *Teleport) Analyze() tea.Msg {
	// Example: Set up targets (this should be based on your actual logic)
	teleport.targets = []Target{
		{Line: 10, Col: 5, Weight: 1},
		{Line: 20, Col: 15, Weight: 2},
		// Add more targets as needed
	}

	// Assign unique codes to targets
	teleport.assignCodes()

	// Return a message or command if needed
	return nil
}

// View returns the teleport overlay as a string
func (teleport *Teleport) View() string {
	if !teleport.active {
		return ""
	}

	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("25")).
		Foreground(lipgloss.Color("220")).
		Bold(true).
		Padding(0, 1)

	// Style for matched characters
	matchedStyle := baseStyle.
		Foreground(lipgloss.Color("212")) // Bright pink for matched characters

	var overlays []string
	for _, target := range teleport.targets {
		displayText := target.Code
		
		// Highlight matched portion if there's input
		if len(teleport.input) > 0 && len(target.Code) >= len(teleport.input) {
			if target.Code[:len(teleport.input)] == teleport.input {
				matched := matchedStyle.Render(target.Code[:len(teleport.input)])
				remaining := baseStyle.Render(target.Code[len(teleport.input):])
				displayText = matched + remaining
			}
		} else {
			displayText = baseStyle.Render(displayText)
		}

		overlays = append(overlays, displayText)
	}

	// Join all overlays with spaces between them
	return lipgloss.JoinHorizontal(lipgloss.Left, overlays...)
}

// assignCodes generates unique codes for all targets
func (teleport *Teleport) assignCodes() {
	// Sort targets by weight
	sort.Slice(teleport.targets, func(i, j int) bool {
		return teleport.targets[i].Weight > teleport.targets[j].Weight
	})

	// Use easily distinguishable characters
	chars := []rune("asdfjkl;weruioqghty")

	// Assign codes
	for i := range teleport.targets {
		if i < len(chars) {
			teleport.targets[i].Code = string(chars[i])
		} else {
			// Generate two-character codes for remaining targets
			first := chars[(i/len(chars))%len(chars)]
			second := chars[i%len(chars)]
			teleport.targets[i].Code = string(first) + string(second)
		}
	}
}
