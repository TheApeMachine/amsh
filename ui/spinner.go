package ui

import (
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

const (
	charCyclingFPS  = time.Second / 22
	colorCycleFPS   = time.Second / 5
	maxCyclingChars = 120
)

var charRunes = []rune("0123456789abcdefABCDEF~!@#$£€%^&*()+=_")

type charState int

const (
	charInitialState charState = iota
	charCyclingState
	charEndOfLifeState
)

// cyclingChar is a single animated character.
type cyclingChar struct {
	finalValue   rune // if < 0 cycle forever
	currentValue rune
	initialDelay time.Duration
	lifetime     time.Duration
}

func (c cyclingChar) randomRune() rune {
	return (charRunes)[rand.Intn(len(charRunes))] //nolint:gosec
}

func (c cyclingChar) state(start time.Time) charState {
	now := time.Now()
	if now.Before(start.Add(c.initialDelay)) {
		return charInitialState
	}
	if c.finalValue > 0 && now.After(start.Add(c.initialDelay)) {
		return charEndOfLifeState
	}
	return charCyclingState
}

type stepCharsMsg struct{}

func stepChars() tea.Cmd {
	return tea.Tick(charCyclingFPS, func(time.Time) tea.Msg {
		return stepCharsMsg{}
	})
}

type colorCycleMsg struct{}

func cycleColors() tea.Cmd {
	return tea.Tick(colorCycleFPS, func(time.Time) tea.Msg {
		return colorCycleMsg{}
	})
}

// Spinner is the model that manages the animation that displays while the
// output is being generated.
type Spinner struct {
	start           time.Time
	cyclingChars    []cyclingChar
	labelChars      []cyclingChar
	ramp            []lipgloss.Style
	label           []rune
	ellipsis        spinner.Model
	ellipsisStarted bool
	styles          lipgloss.Style
}

func NewSpinner(cyclingCharsSize uint, label string, r *lipgloss.Renderer, s lipgloss.Style) *Spinner {
	// #nosec G115
	n := int(cyclingCharsSize)
	if n > maxCyclingChars {
		n = maxCyclingChars
	}

	gap := " "
	if n == 0 {
		gap = ""
	}

	c := Spinner{
		start:    time.Now(),
		label:    []rune(gap + label),
		ellipsis: spinner.New(spinner.WithSpinner(spinner.Ellipsis)),
		styles:   s,
	}

	// If we're in truecolor mode (and there are enough cycling characters)
	// color the cycling characters with a gradient ramp.
	const minRampSize = 3
	if n >= minRampSize && r.ColorProfile() == termenv.TrueColor {
		// Note: double capacity for color cycling as we'll need to reverse and
		// append the ramp for seamless transitions.
		c.ramp = make([]lipgloss.Style, n, n*2) //nolint:mnd
		ramp := makeGradientRamp(n)
		for i, color := range ramp {
			c.ramp[i] = r.NewStyle().Foreground(color)
		}
		c.ramp = append(c.ramp, reverse(c.ramp)...) // reverse and append for color cycling
	}

	makeDelay := func(a int32, b time.Duration) time.Duration {
		return time.Duration(rand.Int31n(a)) * (time.Millisecond * b) //nolint:gosec
	}

	makeInitialDelay := func() time.Duration {
		return makeDelay(8, 60) //nolint:mnd
	}

	// Initial characters that cycle forever.
	c.cyclingChars = make([]cyclingChar, n)

	for i := 0; i < n; i++ {
		c.cyclingChars[i] = cyclingChar{
			finalValue:   -1, // cycle forever
			initialDelay: makeInitialDelay(),
		}
	}

	// Label text that only cycles for a little while.
	c.labelChars = make([]cyclingChar, len(c.label))

	for i, r := range c.label {
		c.labelChars[i] = cyclingChar{
			finalValue:   r,
			initialDelay: makeInitialDelay(),
			lifetime:     makeDelay(5, 180), //nolint:mnd
		}
	}

	return &c
}

// Init initializes the animation.
func (s *Spinner) Init() tea.Cmd {
	return tea.Batch(stepChars(), cycleColors())
}

// Update handles messages.
func (s *Spinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case stepCharsMsg:
		s.updateChars(&s.cyclingChars)
		s.updateChars(&s.labelChars)

		if !s.ellipsisStarted {
			var eol int
			for _, c := range s.labelChars {
				if c.state(s.start) == charEndOfLifeState {
					eol++
				}
			}
			if eol == len(s.label) {
				// If our entire label has reached end of life, start the
				// ellipsis "spinner" after a short pause.
				s.ellipsisStarted = true
				cmd = tea.Tick(time.Millisecond*220, func(time.Time) tea.Msg { //nolint:mnd
					return s.ellipsis.Tick()
				})
			}
		}

		return s, tea.Batch(stepChars(), cmd)
	case colorCycleMsg:
		const minColorCycleSize = 2
		if len(s.ramp) < minColorCycleSize {
			return s, nil
		}
		s.ramp = append(s.ramp[1:], s.ramp[0])
		return s, cycleColors()
	case spinner.TickMsg:
		var cmd tea.Cmd
		s.ellipsis, cmd = s.ellipsis.Update(msg)
		return s, cmd
	default:
		return s, nil
	}
}

func (s *Spinner) updateChars(chars *[]cyclingChar) {
	for i, c := range *chars {
		switch c.state(s.start) {
		case charInitialState:
			(*chars)[i].currentValue = '.'
		case charCyclingState:
			(*chars)[i].currentValue = c.randomRune()
		case charEndOfLifeState:
			(*chars)[i].currentValue = c.finalValue
		}
	}
}

// View renders the animation.
func (s *Spinner) View() string {
	var b strings.Builder

	for i, c := range s.cyclingChars {
		if len(s.ramp) > i {
			b.WriteString(s.ramp[i].Render(string(c.currentValue)))
			continue
		}
		b.WriteRune(c.currentValue)
	}

	for _, c := range s.labelChars {
		b.WriteRune(c.currentValue)
	}

	return b.String() + s.ellipsis.View()
}

func makeGradientRamp(length int) []lipgloss.Color {
	const startColor = "#F967DC"
	const endColor = "#6B50FF"
	var (
		c        = make([]lipgloss.Color, length)
		start, _ = colorful.Hex(startColor)
		end, _   = colorful.Hex(endColor)
	)
	for i := 0; i < length; i++ {
		step := start.BlendLuv(end, float64(i)/float64(length))
		c[i] = lipgloss.Color(step.Hex())
	}
	return c
}

func MakeGradientText(baseStyle lipgloss.Style, str string) string {
	const minSize = 3
	if len(str) < minSize {
		return str
	}
	b := strings.Builder{}
	runes := []rune(str)
	for i, c := range makeGradientRamp(len(str)) {
		b.WriteString(baseStyle.Foreground(c).Render(string(runes[i])))
	}
	return b.String()
}

func reverse[T any](in []T) []T {
	out := make([]T, len(in))
	copy(out, in[:])
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
