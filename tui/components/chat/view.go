package chat

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/theapemachine/amsh/tui/core"
	"golang.org/x/exp/rand"
)

type ChatWindow struct {
	x                 int
	y                 int
	width             int
	height            int
	Active            bool
	buffer            *core.Buffer
	underlyingContent []string
	inputBuffer       string
}

func NewChatWindow(width, height int, buffer *core.Buffer) *ChatWindow {
	cw := &ChatWindow{
		width:  int(float64(width) * 0.8),
		height: int(float64(height) * 0.3),
		x:      (width - int(float64(width)*0.8)) / 2,
		y:      (height - int(float64(height)*0.3)) / 2,
		Active: false,
		buffer: buffer,
	}
	return cw
}

func (cw *ChatWindow) SaveUnderlyingContent() {
	cw.underlyingContent = make([]string, cw.height)
	for i := 0; i < cw.height; i++ {
		lineIdx := cw.y + i
		if lineIdx < len(cw.buffer.Data) {
			cw.underlyingContent[i] = string(cw.buffer.Data[lineIdx])
		} else {
			cw.underlyingContent[i] = ""
		}
	}
	cw.SaveCurrentCursorPosition()
}

func (cw *ChatWindow) RestoreUnderlyingContent() {
	for i, content := range cw.underlyingContent {
		lineIdx := cw.y + i
		if lineIdx < len(cw.buffer.Data) {
			cw.buffer.Data[lineIdx] = []rune(content)
		}
		fmt.Printf("\033[%d;1H", lineIdx+1)
		fmt.Print("\033[K")
		fmt.Print(content)
	}
	cw.RestoreCursorPosition()
	cw.flushStdout()
}

func (cw *ChatWindow) Draw() {
	// Draw the chat window border with double lines
	fmt.Printf("\033[%d;%dH", cw.y, cw.x)

	fmt.Print("╔" + strings.Repeat("═", cw.width-2) + "╗")

	for i := 1; i < cw.height-1; i++ {
		fmt.Printf("\033[%d;%dH", cw.y+i, cw.x)
		fmt.Print("║" + strings.Repeat(" ", cw.width-2) + "║")
	}

	fmt.Printf("\033[%d;%dH", cw.y+cw.height-1, cw.x)
	fmt.Print("╚" + strings.Repeat("═", cw.width-2) + "╝")

	// Draw title
	title := " Chat Window "
	titleX := cw.x + (cw.width-len(title))/2
	fmt.Printf("\033[%d;%dH", cw.y, titleX)
	fmt.Printf("\033[1m%s\033[0m", title) // Bold text

	err := cw.DrawLogo("tui/logo.ans")

	if err != nil {
		fmt.Printf("Error drawing logo: %v", err)
	}

	cw.ShowAdvancedAnimations()
	cw.DrawMessages()
	cw.SetCursorPosition()
	cw.flushStdout()
}

func (cw *ChatWindow) flushStdout() {
	os.Stdout.Sync()
}

func (cw *ChatWindow) AddMessage(msg string) {
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := fmt.Sprintf("\033[33m%s\033[0m %s", timestamp, msg)
	cw.buffer.Data = append(cw.buffer.Data, []rune(formattedMsg))
	if len(cw.buffer.Data) > cw.height-3 {
		cw.buffer.Data = cw.buffer.Data[1:]
	}
	cw.DrawMessages()
}

func (cw *ChatWindow) DrawMessages() {
	for i, msg := range cw.buffer.Data {
		fmt.Printf("\033[%d;%dH", cw.y+1+i, cw.x+1)
		fmt.Printf("%-*s", cw.width-2, string(msg))
	}
	cw.flushStdout()
}

func (cw *ChatWindow) UpdateInputDisplay(input string) {
	cw.inputBuffer = input
	fmt.Printf("\033[%d;%dH", cw.y+cw.height-2, cw.x+1)
	fmt.Print(strings.Repeat(" ", cw.width-2))
	fmt.Printf("\033[%d;%dH", cw.y+cw.height-2, cw.x+1)
	fmt.Printf("\033[36mInput:\033[0m %s", input)
	cw.flushStdout()
}

func (cw *ChatWindow) SetCursorPosition() {
	prefix := "Input: "
	xPosition := cw.x + len(prefix) + 1
	cw.buffer.Cursor.Move(xPosition, cw.y+cw.height-2)
	fmt.Printf("\033[%d;%dH", cw.y+cw.height-2, xPosition)
	cw.flushStdout()
}

func (cw *ChatWindow) SaveCurrentCursorPosition() {
	cw.buffer.Cursor.Save()
}

func (cw *ChatWindow) RestoreCursorPosition() {
	cw.buffer.Cursor.Restore()
}

func (cw *ChatWindow) Activate() {
	cw.Active = true
	cw.SaveUnderlyingContent()
	cw.Draw()
}

func (cw *ChatWindow) Deactivate() {
	cw.Active = false
	cw.RestoreUnderlyingContent()
}

func (cw *ChatWindow) HandleInput(ch rune) {
	switch ch {
	case 13: // Enter key
		if cw.inputBuffer != "" {
			cw.AddMessage(cw.inputBuffer)
			cw.inputBuffer = ""
			cw.UpdateInputDisplay("")
		}
	case 127: // Backspace
		if len(cw.inputBuffer) > 0 {
			cw.inputBuffer = cw.inputBuffer[:len(cw.inputBuffer)-1]
			cw.UpdateInputDisplay(cw.inputBuffer)
		}
	default:
		cw.inputBuffer += string(ch)
		cw.UpdateInputDisplay(cw.inputBuffer)
	}
}

func (cw *ChatWindow) DrawLogo(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	maxWidth := 0

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Calculate starting position to center the logo
	startX := cw.x + cw.width - 5
	startY := cw.y + 2 // Leave some space at the top

	// Draw each line of the logo
	for i, line := range lines {
		fmt.Printf("\033[%d;%dH", startY+i, startX)
		fmt.Print(line)
	}

	cw.flushStdout()
	return nil
}

func (cw *ChatWindow) AnimateLoading(ctx context.Context, x, y int, message string) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-ctx.Done():
			// Clear the animation
			fmt.Printf("\033[%d;%dH%s", y, x, strings.Repeat(" ", len(message)+3))
			cw.flushStdout()
			return
		default:
			frame := frames[i%len(frames)]
			fmt.Printf("\033[%d;%dH\033[36m%s\033[0m %s", y, x, frame, message)
			cw.flushStdout()
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}

func (cw *ChatWindow) AnimateText(ctx context.Context, x, y int, text string, delay time.Duration) {
	for i := 0; i <= len(text); i++ {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Printf("\033[%d;%dH%s", y, x, text[:i])
			cw.flushStdout()
			time.Sleep(delay)
		}
	}
}

func (cw *ChatWindow) AnimateBorder(ctx context.Context, duration time.Duration) {
	colors := []string{"\033[31m", "\033[33m", "\033[32m", "\033[36m", "\033[34m", "\033[35m"}
	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			// Reset border color
			cw.Draw()
			return
		default:
			elapsedTime := time.Since(startTime)
			if elapsedTime >= duration {
				cw.Draw()
				return
			}
			colorIndex := int(elapsedTime.Milliseconds()/100) % len(colors)
			color := colors[colorIndex]

			// Draw top and bottom borders
			fmt.Printf("\033[%d;%dH%s%s\033[0m", cw.y, cw.x, color, strings.Repeat("═", cw.width))
			fmt.Printf("\033[%d;%dH%s%s\033[0m", cw.y+cw.height-1, cw.x, color, strings.Repeat("═", cw.width))

			// Draw left and right borders
			for i := 1; i < cw.height-1; i++ {
				fmt.Printf("\033[%d;%dH%s║\033[0m", cw.y+i, cw.x, color)
				fmt.Printf("\033[%d;%dH%s║\033[0m", cw.y+i, cw.x+cw.width-1, color)
			}

			cw.flushStdout()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Example usage in the ChatWindow struct
func (cw *ChatWindow) ShowLoadingAnimation(message string) {
	ctx, cancel := context.WithCancel(context.Background())
	go cw.AnimateLoading(ctx, cw.x+2, cw.y+cw.height-3, message)

	// Simulate some work
	time.Sleep(5 * time.Second)

	// Stop the animation
	cancel()
}

func (cw *ChatWindow) ShowWelcomeMessage() {
	ctx, cancel := context.WithCancel(context.Background())
	go cw.AnimateText(ctx, cw.x+2, cw.y+2, "Welcome to the Chat Window!", 50*time.Millisecond)

	// Let the animation complete
	time.Sleep(2 * time.Second)

	// Stop the animation
	cancel()
}

func (cw *ChatWindow) ShowColorfulBorder() {
	ctx, cancel := context.WithCancel(context.Background())
	go cw.AnimateBorder(ctx, 5*time.Second)

	// Let the animation run for its duration
	time.Sleep(5 * time.Second)

	// Stop the animation
	cancel()
}

func (cw *ChatWindow) AnimateSlideIn(ctx context.Context, message string) {
	messageWidth := len(message)
	for i := 0; i <= messageWidth; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			padding := strings.Repeat(" ", messageWidth-i)
			fmt.Printf("\033[%d;%dH%s%s", cw.y+cw.height/2, cw.x, padding, message[:i])
			cw.flushStdout()
			time.Sleep(30 * time.Millisecond)
		}
	}
}

func (cw *ChatWindow) AnimateFadeIn(ctx context.Context, message string) {
	colors := []string{"\033[38;5;232m", "\033[38;5;236m", "\033[38;5;240m", "\033[38;5;244m", "\033[38;5;248m", "\033[38;5;252m", "\033[0m"}
	for _, color := range colors {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Printf("\033[%d;%dH%s%s\033[0m", cw.y+cw.height/2, cw.x, color, message)
			cw.flushStdout()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (cw *ChatWindow) AnimateProgressBar(ctx context.Context, duration time.Duration) {
	width := cw.width - 4
	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			elapsed := time.Since(startTime)
			if elapsed >= duration {
				fmt.Printf("\033[%d;%dH[%s]", cw.y+cw.height-3, cw.x+2, strings.Repeat("=", width))
				cw.flushStdout()
				return
			}
			progress := float64(elapsed) / float64(duration)
			filled := int(progress * float64(width))
			fmt.Printf("\033[%d;%dH[%s%s]", cw.y+cw.height-3, cw.x+2, strings.Repeat("=", filled), strings.Repeat(" ", width-filled))
			cw.flushStdout()
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (cw *ChatWindow) AnimateParticles(ctx context.Context, duration time.Duration) {
	type Particle struct {
		x, y   float64
		dx, dy float64
		char   rune
		color  string
	}

	particles := make([]Particle, 20)
	for i := range particles {
		particles[i] = Particle{
			x:     float64(cw.width / 2),
			y:     float64(cw.height / 2),
			dx:    (rand.Float64() - 0.5) * 2,
			dy:    (rand.Float64() - 0.5) * 2,
			char:  []rune("*+.◦°")[rand.Intn(5)],
			color: []string{"\033[31m", "\033[33m", "\033[32m", "\033[36m", "\033[34m", "\033[35m"}[rand.Intn(6)],
		}
	}

	startTime := time.Now()
	for time.Since(startTime) < duration {
		select {
		case <-ctx.Done():
			return
		default:
			// Clear previous frame
			for y := 0; y < cw.height; y++ {
				fmt.Printf("\033[%d;%dH%s", cw.y+y, cw.x, strings.Repeat(" ", cw.width))
			}

			// Update and draw particles
			for i := range particles {
				particles[i].x += particles[i].dx
				particles[i].y += particles[i].dy

				// Bounce off walls
				if particles[i].x < 0 || particles[i].x >= float64(cw.width) {
					particles[i].dx *= -1
				}
				if particles[i].y < 0 || particles[i].y >= float64(cw.height) {
					particles[i].dy *= -1
				}

				// Draw particle
				fmt.Printf("\033[%d;%dH%s%c\033[0m",
					cw.y+int(particles[i].y),
					cw.x+int(particles[i].x),
					particles[i].color,
					particles[i].char)
			}

			cw.flushStdout()
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (cw *ChatWindow) AnimateWaveText(ctx context.Context, text string, cycles int) {
	width := cw.width - 2
	padding := strings.Repeat(" ", (width-len(text))/2)
	text = padding + text + padding

	for cycle := 0; cycle < cycles; cycle++ {
		for i := 0; i < 360; i += 15 {
			select {
			case <-ctx.Done():
				return
			default:
				var wavedText strings.Builder
				for j, char := range text {
					y := int(math.Sin(float64(i+j*10)*math.Pi/180) * 2)
					wavedText.WriteString(fmt.Sprintf("\033[%d;%dH%c", cw.y+cw.height/2+y, cw.x+j, char))
				}
				fmt.Print(wavedText.String())
				cw.flushStdout()
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

// Example usage in the ChatWindow struct
func (cw *ChatWindow) ShowAdvancedAnimations() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cw.AnimateSlideIn(ctx, "Welcome to Advanced Animations!")
	time.Sleep(1 * time.Second)

	cw.AnimateFadeIn(ctx, "Fading in slowly...")
	time.Sleep(1 * time.Second)

	go cw.AnimateProgressBar(ctx, 5*time.Second)
	time.Sleep(5 * time.Second)

	cw.AnimateParticles(ctx, 5*time.Second)

	cw.AnimateWaveText(ctx, "Wavy Text Animation", 3)
}
