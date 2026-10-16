package ui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Braille animation frames for smooth high-framerate rendering
var BrailleFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var setupSignalOnce sync.Once

func ensureSignalHandler() {
	setupSignalOnce.Do(func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Print("\033[?25h\r\033[K")
			os.Exit(130)
		}()
	})
}

func stripANSI(str string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range str {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func truncatePlain(s string, maxLen int) string {
	if maxLen <= 3 {
		if maxLen <= 0 {
			return ""
		}
		runes := []rune(s)
		if len(runes) > maxLen {
			return string(runes[:maxLen])
		}
		return s
	}
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return s
}

// Spinner provides a clean, purple-themed braille spinner without brackets or bold.
type Spinner struct {
	message   string
	stop      chan struct{}
	done      chan struct{}
	active    bool
	mu        sync.Mutex
	startTime time.Time
}

// NewSpinner constructs a new Spinner initialized with braille frames.
func NewSpinner(msg string) *Spinner {
	ensureSignalHandler()
	s := &Spinner{
		message:   msg,
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
		startTime: time.Now(),
	}
	s.Start()
	return s
}

// Start begins the animation loop in a background goroutine.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	// If NO_COLOR or dumb terminal, log message plainly without animation
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		Info(s.message)
		close(s.done)
		return
	}

	// Hide cursor during active animation
	fmt.Print("\033[?25l")
	_ = os.Stdout.Sync()

	go func() {
		defer close(s.done)
		ticker := time.NewTicker(75 * time.Millisecond)
		defer ticker.Stop()
		i := 0

		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K")
				_ = os.Stdout.Sync()
				return
			case <-ticker.C:
				s.mu.Lock()
				msg := s.message
				s.mu.Unlock()

				cols := TerminalWidth()
				if cols <= 0 {
					cols = 80
				}

				frame := BrailleFrames[i%len(BrailleFrames)]
				prefix := fmt.Sprintf("%s ", BrightPurple(frame))
				prefixLen := 2
				avail := cols - prefixLen - 2
				if avail < 4 {
					avail = 4
				}

				displayMsg := msg
				plainMsg := stripANSI(msg)
				if len(plainMsg) > avail {
					displayMsg = truncatePlain(msg, avail)
				}

				fmt.Printf("\r\033[K%s%s", prefix, displayMsg)
				_ = os.Stdout.Sync()
				i++
			}
		}
	}()
}

// UpdateMessage dynamically updates the text displayed next to the spinner.
func (s *Spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	s.message = msg
	s.mu.Unlock()
}

// StopSuccess halts the spinner and prints a green checkmark success status line.
func (s *Spinner) StopSuccess(msg string) {
	s.stopAnimation()
	Success(msg)
}

// StopFail halts the spinner and prints a red cross error status line.
func (s *Spinner) StopFail(msg string) {
	s.stopAnimation()
	Fail(msg)
}

// StopWarn halts the spinner and prints a yellow warning status line.
func (s *Spinner) StopWarn(msg string) {
	s.stopAnimation()
	Warn(msg)
}

// StopInfo halts the spinner and prints a purple info bullet status line.
func (s *Spinner) StopInfo(msg string) {
	s.stopAnimation()
	Info(msg)
}

func (s *Spinner) stopAnimation() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()

	close(s.stop)
	<-s.done

	// Restore cursor and clear the animation line
	fmt.Print("\r\033[K\033[?25h")
	_ = os.Stdout.Sync()
}

// RenderProgressBar returns a clean purple progress bar string matching vtermux style.
func RenderProgressBar(current, total int64, width int) string {
	if total <= 0 {
		return ""
	}
	if current > total {
		current = total
	}
	percent := float64(current) / float64(total) * 100
	completed := int(float64(width) * float64(current) / float64(total))
	if completed > width {
		completed = width
	}
	bar := Purple(strings.Repeat("━", completed)) + GrayText(strings.Repeat("─", width-completed))

	return fmt.Sprintf("[%s] %s%3.0f%%%s", bar, ColorGreen, percent, ColorReset)
}
