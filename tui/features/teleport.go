package features

import (
	"sort"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/theapemachine/amsh/tui/core"
)

type Target struct {
	Line   int
	Col    int
	Code   string
	Weight int // Priority weight for sorting targets
}

type Teleport struct {
	active  bool
	targets []Target
	input   string // Current input sequence
}

func NewTeleport() *Teleport {
	return &Teleport{
		active:  false,
		targets: make([]Target, 0),
	}
}

// Toggle activates or deactivates teleport mode
func (teleport *Teleport) Toggle() {
	teleport.active = !teleport.active
	teleport.input = ""
}

// IsActive returns whether teleport mode is active
func (teleport *Teleport) IsActive() bool {
	return teleport.active
}

// AddInput adds a character to the input sequence and returns true if it matches a target
func (teleport *Teleport) AddInput(char rune) (matched bool, target Target) {
	teleport.input += string(char)

	// Check for exact match
	for _, t := range teleport.targets {
		if t.Code == teleport.input {
			return true, t
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

	// If no targets match the input, reset
	if !matching {
		teleport.input = ""
	}

	return false, Target{}
}

// Analyze scans the buffer for jump targets
func (teleport *Teleport) Analyze(buffer *core.Buffer, cursorLine, cursorCol int) {
	teleport.targets = make([]Target, 0)

	// Characters we consider important for navigation
	important := map[rune]bool{
		'{': true, '}': true, '(': true, ')': true,
		'[': true, ']': true, '"': true, '\'': true,
		'.': true, ',': true, ':': true, ';': true,
	}

	for lineNum := 0; lineNum < buffer.LineCount(); lineNum++ {
		line := buffer.GetLine(lineNum)
		wasSpace := true // Track if previous char was space

		for col, char := range line {
			weight := 0
			isTarget := false

			// Word start (after space)
			if wasSpace && unicode.IsLetter(char) {
				weight = 100
				isTarget = true
			}

			// Important symbols
			if important[char] {
				weight = 80
				isTarget = true
			}

			// Capital letters (likely start of important words)
			if unicode.IsUpper(char) {
				weight = 90
				isTarget = true
			}

			if isTarget {
				// Adjust weight based on distance from cursor
				distance := abs(lineNum-cursorLine) + abs(col-cursorCol)
				weight -= distance / 2

				teleport.targets = append(teleport.targets, Target{
					Line:   lineNum,
					Col:    col,
					Weight: weight,
				})
			}

			wasSpace = unicode.IsSpace(char)
		}
	}

	// Sort and assign codes
	teleport.assignCodes()
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

// Render draws the teleport targets on the screen
func (teleport *Teleport) Render(screen tcell.Screen) {
	if !teleport.active {
		return
	}

	style := tcell.StyleDefault.
		Background(tcell.ColorDarkBlue).
		Foreground(tcell.ColorYellow).
		Bold(true)

	for _, target := range teleport.targets {
		// Highlight matching prefix if there's input
		matchStyle := style
		if len(teleport.input) > 0 && len(target.Code) >= len(teleport.input) &&
			target.Code[:len(teleport.input)] == teleport.input {
			matchStyle = matchStyle.Background(tcell.ColorGreen)
		}

		// Draw the target code
		for i, r := range target.Code {
			screen.SetContent(
				target.Col+i,
				target.Line,
				r,
				nil,
				matchStyle,
			)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
