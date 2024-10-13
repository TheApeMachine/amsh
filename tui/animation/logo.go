package animation

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

type Logo struct {
	Pattern [][]rune
	Width   int
	Height  int
}

func LoadLogoFromFile(filename string) (*Logo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var pattern [][]rune
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " ") // Remove trailing spaces
		pattern = append(pattern, []rune(line))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	if len(pattern) == 0 {
		return nil, fmt.Errorf("logo file is empty")
	}

	width := len(pattern[0])
	for _, line := range pattern {
		if len(line) != width {
			return nil, fmt.Errorf("inconsistent line lengths in logo file")
		}
	}

	return &Logo{
		Pattern: pattern,
		Width:   width,
		Height:  len(pattern),
	}, nil
}

type LogoBuildUp struct {
	Logo *Logo
}

func (l LogoBuildUp) Update(canvas *Canvas, progress float64) {
	centerX := (canvas.Width - l.Logo.Width) / 2
	centerY := (canvas.Height - l.Logo.Height) / 2
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789@#$%^&*()_+-=[]{}|;:,.<>?"

	for y := 0; y < l.Logo.Height; y++ {
		for x := 0; x < l.Logo.Width; x++ {
			if l.Logo.Pattern[y][x] != ' ' {
				threshold := 1.0 - float64(y)/float64(l.Logo.Height)
				if progress > threshold {
					canvas.Set(centerX+x, centerY+y, l.Logo.Pattern[y][x])
				} else if progress > threshold-0.1 {
					canvas.Set(centerX+x, centerY+y, rune(chars[rand.Intn(len(chars))]))
				}
			}
		}
	}
}

type RainbowCycle struct {
	Logo *Logo
}

func (r RainbowCycle) Update(canvas *Canvas, progress float64) {
	colors := []string{"\033[31m", "\033[33m", "\033[32m", "\033[36m", "\033[34m", "\033[35m"}
	centerX := (canvas.Width - r.Logo.Width) / 2
	centerY := (canvas.Height - r.Logo.Height) / 2

	for y := 0; y < r.Logo.Height; y++ {
		for x := 0; x < r.Logo.Width; x++ {
			if r.Logo.Pattern[y][x] != ' ' {
				colorIndex := int(progress*6+float64(x+y)/5) % len(colors)
				canvas.SetColored(centerX+x, centerY+y, r.Logo.Pattern[y][x], colors[colorIndex])
			}
		}
	}
}

type Explosion struct {
	Logo *Logo
}

func (e Explosion) Update(canvas *Canvas, progress float64) {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789@#$%^&*()_+-=[]{}|;:,.<>?"
	colors := []string{"\033[31m", "\033[33m", "\033[32m", "\033[36m", "\033[34m", "\033[35m"}
	centerX := canvas.Width / 2
	centerY := canvas.Height / 2

	for y := 0; y < canvas.Height; y++ {
		for x := 0; x < canvas.Width; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			distance := math.Sqrt(dx*dx + dy*dy)
			if distance < progress*float64(canvas.Width/2) {
				char := rune(chars[rand.Intn(len(chars))])
				color := colors[rand.Intn(len(colors))]
				canvas.SetColored(x, y, char, color)
			}
		}
	}
}

func Render() {
	logo, err := LoadLogoFromFile("tui/logo.ans")
	if err != nil {
		fmt.Printf("Error loading logo: %v\n", err)
		return
	}

	canvas := NewCanvas(80, 24) // Adjust size as needed
	ctx := context.Background()

	sequence := Sequence{
		&LogoBuildUp{Logo: logo},
		&Composite{
			Animations: []Animation{
				&LogoBuildUp{Logo: logo},
				&RainbowCycle{Logo: logo},
			},
			Duration: 3 * time.Second,
		},
		&Explosion{Logo: logo},
	}

	Animate(ctx, canvas, sequence, 10*time.Second)
}
