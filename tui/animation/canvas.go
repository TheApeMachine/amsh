package animation

import (
	"context"
	"fmt"
	"time"
)

type Canvas struct {
	Width, Height int
	Buffer        [][]struct {
		Rune  rune
		Color string
	}
}

func (c *Canvas) SetColored(x, y int, r rune, color string) {
	if x >= 0 && x < c.Width && y >= 0 && y < c.Height {
		c.Buffer[y][x].Rune = r
		c.Buffer[y][x].Color = color
	}
}

func (c *Canvas) Render() {
	fmt.Print("\033[H\033[2J") // Clear screen and move cursor to top-left
	for _, row := range c.Buffer {
		for _, cell := range row {
			fmt.Print(cell.Color, string(cell.Rune), "\033[0m")
		}
		fmt.Println()
	}
}

func NewCanvas(width, height int) *Canvas {
	buffer := make([][]struct {
		Rune  rune
		Color string
	}, height)
	for i := range buffer {
		buffer[i] = make([]struct {
			Rune  rune
			Color string
		}, width)
		for j := range buffer[i] {
			buffer[i][j] = struct {
				Rune  rune
				Color string
			}{Rune: ' ', Color: "\033[0m"}
		}
	}
	return &Canvas{Width: width, Height: height, Buffer: buffer}
}

func (c *Canvas) Clear() {
	for i := range c.Buffer {
		for j := range c.Buffer[i] {
			c.Buffer[i][j] = struct {
				Rune  rune
				Color string
			}{Rune: ' ', Color: "\033[0m"}
		}
	}
}

func (c *Canvas) Set(x, y int, r rune) {
	if x >= 0 && x < c.Width && y >= 0 && y < c.Height {
		c.Buffer[y][x] = struct {
			Rune  rune
			Color string
		}{Rune: r, Color: "\033[0m"}
	}
}

type Animation interface {
	Update(canvas *Canvas, progress float64)
}

func Animate(ctx context.Context, canvas *Canvas, anim Animation, duration time.Duration) {
	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			elapsed := time.Since(startTime)
			if elapsed >= duration {
				return
			}
			progress := float64(elapsed) / float64(duration)
			canvas.Clear()
			anim.Update(canvas, progress)
			canvas.Render()
			time.Sleep(16 * time.Millisecond) // ~60 FPS
		}
	}
}

type FadeIn struct {
	Text string
}

func (f FadeIn) Update(canvas *Canvas, progress float64) {
	x := (canvas.Width - len(f.Text)) / 2
	y := canvas.Height / 2
	for i, ch := range f.Text {
		if progress > float64(i)/float64(len(f.Text)) {
			canvas.Set(x+i, y, ch)
		}
	}
}
