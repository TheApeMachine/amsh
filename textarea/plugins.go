package textarea

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/logger"
)

var modifier = func(s string) string {
	// no-op
	return s
}

/*
easyMotion works in the following way:
1. It dims the text currently visible in the textarea
2. It highlights the first n characters of each words, such that each word is highlighted with the minimal amount of characters needed for each highlight to be unique.
3. It waits for a secondary input and moves the cursor to the highlight that matches the secondary input.
4. If the secondary input is esc, the plugin is cancelled.
5. If the secondary input matches a highlighted character, the cursor moves to the next occurence of that character.
*/
func (model *Model) easyMotion() tea.Cmd {
	logger.Debug("textarea.easyMotion()")

	model.plugin = func(s string) string {
		return model.highlight(s)
	}

	return nil
}

/*
highlight
*/
func (model *Model) highlight(s string) string {
	// Define a lookup table with only the letters we care about (a-z and A-Z)
	letters := map[rune]bool{}
	for ch := 'a'; ch <= 'z'; ch++ {
		letters[ch] = true
	}
	for ch := 'A'; ch <= 'Z'; ch++ {
		letters[ch] = true
	}

	var result strings.Builder
	inEscape := false     // Track if we're inside an ANSI escape sequence
	highlightedCount := 0 // Track how many letters we've highlighted

	for i := 0; i < len(s); i++ {
		char := rune(s[i])

		// Detect ANSI escape sequence start
		if char == '\033' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
		}

		// If inside an escape sequence, skip till we find the end 'm'
		if inEscape {
			result.WriteByte(s[i])
			if char == 'm' { // End of ANSI escape sequence
				inEscape = false
			}
			continue
		}

		// Skip spaces and other characters that are not letters
		if !letters[char] {
			result.WriteByte(byte(char))
			continue
		}

		// If the character is a letter and we haven't highlighted enough, highlight it
		if highlightedCount < 1 {
			styled := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(string(char))
			result.WriteString(styled)
			highlightedCount++
		} else {
			// Just append the character as is
			result.WriteByte(byte(char))
		}
	}

	return result.String()
}
