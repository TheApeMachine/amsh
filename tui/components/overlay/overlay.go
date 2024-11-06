package overlay

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/ansi"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/termenv"
)

// getLines extracts lines from the string and returns the lines along with the widest line length.
func getLines(s string) ([]string, int) {
	lines := strings.Split(s, "\n")
	widest := 0

	for _, line := range lines {
		width := ansi.PrintableRuneWidth(line)
		if width > widest {
			widest = width
		}
	}

	return lines, widest
}

// clamp restricts a value to a given range.
func clamp(v, lower, upper int) int {
	if v < lower {
		return lower
	}
	if v > upper {
		return upper
	}
	return v
}

// PlaceOverlay places the foreground (`fg`) string on top of the background (`bg`) string.
// Optional shadow effect and whitespace styling can be provided.
func PlaceOverlay(x, y int, fg, bg string, shadow bool, opts ...WhitespaceOption) string {
	fgLines, fgWidth := getLines(fg)
	bgLines, bgWidth := getLines(bg)

	if shadow {
		fg = addShadow(fg, fgWidth, len(fgLines))
		fgLines, fgWidth = getLines(fg)
	}

	bgHeight := len(bgLines)
	fgHeight := len(fgLines)

	// Clamp coordinates to ensure fg is placed within bg bounds.
	x = clamp(x, 0, bgWidth-fgWidth)
	y = clamp(y, 0, bgHeight-fgHeight)

	ws := applyWhitespaceOptions(opts)

	var b strings.Builder
	for i, bgLine := range bgLines {
		if i > 0 {
			b.WriteByte('\n')
		}

		if i < y || i >= y+fgHeight {
			b.WriteString(bgLine) // Background only, no foreground overlay for this line.
			continue
		}

		b.WriteString(overlayLine(x, fgLines[i-y], bgLine, ws))
	}

	return b.String()
}

// addShadow adds a shadow effect to the foreground string.
func addShadow(fg string, fgWidth, fgHeight int) string {
	var shadowBg strings.Builder
	shadowChar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333")).
		Render("░")

	// First line is blank, followed by shadow lines.
	shadowBg.WriteString(" " + strings.Repeat(" ", fgWidth) + "\n")
	for i := 1; i <= fgHeight; i++ {
		shadowBg.WriteString(" " + strings.Repeat(shadowChar, fgWidth) + "\n")
	}

	return PlaceOverlay(0, 0, fg, shadowBg.String(), false)
}

// overlayLine overlays the foreground line onto the background line at a given x position.
func overlayLine(x int, fgLine, bgLine string, ws *whitespace) string {
	var b strings.Builder

	// Write left part of the background up to the overlay position.
	left := truncate.String(bgLine, uint(x))
	b.WriteString(left)

	// Fill any missing space if the overlay starts beyond the left part.
	pos := ansi.PrintableRuneWidth(left)
	if pos < x {
		b.WriteString(ws.render(x - pos))
	}

	// Write the foreground line.
	b.WriteString(fgLine)
	pos += ansi.PrintableRuneWidth(fgLine)

	// Write the remaining part of the background line, if any.
	right := cutLeft(bgLine, pos)
	rightWidth := ansi.PrintableRuneWidth(right)
	if rightWidth < ansi.PrintableRuneWidth(bgLine)-pos {
		b.WriteString(ws.render(ansi.PrintableRuneWidth(bgLine) - rightWidth - pos))
	}
	b.WriteString(right)

	return b.String()
}

// cutLeft removes a given number of printable characters from the left of the string.
func cutLeft(s string, cutWidth int) string {
	var (
		pos       int
		isAnsi    bool
		ansiBuf   strings.Builder
		resultBuf strings.Builder
	)

	for _, c := range s {
		// Handle ANSI escape sequences.
		if c == ansi.Marker || isAnsi {
			isAnsi = true
			ansiBuf.WriteRune(c)
			if ansi.IsTerminator(c) {
				isAnsi = false
				// Reset ANSI buffer if it's the reset sequence.
				if strings.HasSuffix(ansiBuf.String(), "[0m") {
					ansiBuf.Reset()
				}
			}
			continue
		}

		// Calculate the rune width only if it's not part of an ANSI escape.
		runeWidth := runewidth.RuneWidth(c)

		// Start adding characters once we reach the required cut width.
		if pos >= cutWidth {
			// Append the current ANSI sequence if we haven't already.
			if ansiBuf.Len() > 0 {
				resultBuf.WriteString(ansiBuf.String())
				ansiBuf.Reset()
			}
			// Append the character.
			resultBuf.WriteRune(c)
		}
		// Update the current position.
		pos += runeWidth
	}

	return resultBuf.String()
}

// applyWhitespaceOptions applies options to configure whitespace rendering.
func applyWhitespaceOptions(opts []WhitespaceOption) *whitespace {
	ws := &whitespace{}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}

// whitespace defines whitespace styling.
type whitespace struct {
	style termenv.Style
	chars string
}

// render generates whitespace of the specified width.
func (w whitespace) render(width int) string {
	if w.chars == "" {
		w.chars = " "
	}

	runes := []rune(w.chars)
	var b strings.Builder

	for i, j := 0, 0; i < width; {
		b.WriteRune(runes[j])
		i += ansi.PrintableRuneWidth(string(runes[j]))
		j = (j + 1) % len(runes)
	}

	// Fill any gaps left by wide runes.
	short := width - ansi.PrintableRuneWidth(b.String())
	if short > 0 {
		b.WriteString(strings.Repeat(" ", short))
	}

	return w.style.Styled(b.String())
}

// WhitespaceOption configures whitespace rendering.
type WhitespaceOption func(*whitespace)
