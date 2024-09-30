package tui

import (
	"context"
	"math"
	"time"

	"github.com/theapemachine/amsh/tui/animation"
	"golang.org/x/exp/rand"
)

const (
	LogoWidth  = 21
	LogoHeight = 20
)

var LogoPattern = [][]rune{
	{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', ' ', ' '},
	{' ', ' ', ' ', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', ' '},
	{' ', ' ', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' '},
	{' ', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' '},
	{' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' '},
	{' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' '},
	{' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█'},
	{' ', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█'},
	{'█', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█'},
	{'█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█'},
	{'█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', '█', '█', '█', ' ', ' ', ' ', ' ', '█', '█', '█', '█'},
	{'█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', '█', '█', '█'},
	{'█', '█', '█', '█', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', '█', '█'},
	{'█', '█', '█', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' ', '█'},
	{'█', '█', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' ', ' '},
	{'█', ' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' ', ' '},
	{' ', ' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' ', ' '},
	{' ', ' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', ' '},
	{' ', ' ', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█'},
	{'█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█'},
}

type LogoBuildUp struct{}

func (l LogoBuildUp) Update(canvas *animation.Canvas, progress float64) {
	centerX := (canvas.Width - LogoWidth) / 2
	centerY := (canvas.Height - LogoHeight) / 2
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789@#$%^&*()_+-=[]{}|;:,.<>?"

	for y := 0; y < LogoHeight; y++ {
		for x := 0; x < LogoWidth; x++ {
			if LogoPattern[y][x] == '█' {
				threshold := 1.0 - float64(y)/float64(LogoHeight)
				if progress > threshold {
					canvas.Set(centerX+x, centerY+y, '█')
				} else if progress > threshold-0.1 {
					canvas.Set(centerX+x, centerY+y, rune(chars[rand.Intn(len(chars))]))
				}
			}
		}
	}
}

type RainbowCycle struct{}

func (r RainbowCycle) Update(canvas *animation.Canvas, progress float64) {
	colors := []string{"\033[31m", "\033[33m", "\033[32m", "\033[36m", "\033[34m", "\033[35m"}
	centerX := (canvas.Width - LogoWidth) / 2
	centerY := (canvas.Height - LogoHeight) / 2

	for y := 0; y < LogoHeight; y++ {
		for x := 0; x < LogoWidth; x++ {
			if LogoPattern[y][x] == '█' {
				colorIndex := int(progress*6+float64(x+y)/5) % len(colors)
				canvas.SetColored(centerX+x, centerY+y, '█', colors[colorIndex])
			}
		}
	}
}

type Explosion struct{}

func (e Explosion) Update(canvas *animation.Canvas, progress float64) {
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
	canvas := animation.NewCanvas(80, 24) // Adjust size as needed
	ctx := context.Background()

	sequence := animation.Sequence{
		&LogoBuildUp{},
		&RainbowCycle{},
		&Explosion{},
	}

	animation.Animate(ctx, canvas, sequence, 10*time.Second)
}
