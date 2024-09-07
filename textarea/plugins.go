package textarea

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/*
easyMotion dims the text currently visible in the textarea, and highlights the first n characters
of each words, such that each word is highlighted with the minimal amount of characters needed for
each highlight to be unique. It then waits for a secondary input and moves the cursor to the
highlight that matches the secondary input.
*/
func (model *Model) easyMotion() tea.Cmd {
	return func() tea.Msg {
		// Get visible content
		visibleContent := model.viewport.View()
		lines := strings.Split(visibleContent, "\n")

		// Generate highlights
		highlights, highlightMap := generateHighlights(lines)

		// Apply highlights to the content
		highlightedContent := applyHighlights(lines, highlights)

		// Update the viewport content with highlighted text
		model.viewport.SetContent(highlightedContent)

		// Wait for user input
		return easyMotionPromptMsg(highlightMap)
	}
}

func generateHighlights(lines []string) (map[string]string, map[string][2]int) {
	highlights := make(map[string]string)
	highlightMap := make(map[string][2]int)
	chars := "abcdefghijklmnopqrstuvwxyz"
	charIndex := 0

	for row, line := range lines {
		words := strings.Fields(line)
		for _, word := range words {
			if len(word) > 0 && !unicode.IsPunct(rune(word[0])) {
				highlight := string(chars[charIndex])
				charIndex = (charIndex + 1) % len(chars)

				if _, exists := highlights[highlight]; !exists {
					highlights[highlight] = word
					highlightMap[highlight] = [2]int{row, strings.Index(line, word)}
				}
			}
		}
	}

	return highlights, highlightMap
}

func applyHighlights(lines []string, highlights map[string]string) string {
	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("0"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	var result strings.Builder

	for _, line := range lines {
		for _, word := range strings.Fields(line) {
			if highlight, exists := highlights[word]; exists {
				result.WriteString(highlightStyle.Render(highlight) + dimStyle.Render(word[1:]) + " ")
			} else {
				result.WriteString(dimStyle.Render(word) + " ")
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

type easyMotionPromptMsg map[string][2]int

func (model *Model) handleEasyMotionInput(msg tea.KeyMsg, highlightMap map[string][2]int) {
	if pos, exists := highlightMap[msg.String()]; exists {
		model.row = pos[0]
		model.col = pos[1]
		model.viewport.GotoTop() // Reset viewport position
		model.SetCursor(model.col)
	}

	// Restore original content
	model.viewport.SetContent(model.Value())
}
