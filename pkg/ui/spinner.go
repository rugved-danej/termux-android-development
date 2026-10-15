package ui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pterm/pterm"
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

	plain := stripANSI(s)
	runes := []rune(plain)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

type Spinner struct {
	message   string
	stop      chan struct{}
	done      chan struct{}
	startTime time.Time
	mu        sync.Mutex
	active    bool
}

// NewSpinner initializes an in-place braille spinner that prevents line wrapping and line multiplications.
func NewSpinner(message string) *Spinner {
	ensureSignalHandler()
	return &Spinner{
		message: message,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// Start launches the animated braille spinner strictly on a single line.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.startTime = time.Now()
	s.mu.Unlock()

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
				st := s.startTime
				s.mu.Unlock()

				cols := pterm.GetTerminalWidth()
				if cols <= 0 {
					cols = 80
				}

				frame := BrailleFrames[i%len(BrailleFrames)]
				elapsed := fmt.Sprintf(" (%.1fs)", time.Since(st).Seconds())
				prefix := fmt.Sprintf("  [%s%s%s] ", Cyan+Bold, frame, Reset)
				timerStr := Gray + elapsed + Reset

				// Prefix plain length: 6 characters ("  [X] ")
				prefixLen := 6
				timerLen := len(elapsed)
				avail := cols - prefixLen - timerLen - 2
				if avail < 4 {
					avail = 4
				}

				displayMsg := msg
				plainMsg := stripANSI(msg)
				if len(plainMsg) > avail {
					displayMsg = truncatePlain(msg, avail)
				}

				// Always clear line from col 0 to prevent line wrapping & multiplications
				fmt.Printf("\r\033[K%s%s%s", prefix, displayMsg, timerStr)
				_ = os.Stdout.Sync()
				i++
			}
		}
	}()
}

// UpdateMessage dynamically updates the text displayed next to the spinner.
// It is thread-safe and non-blocking: updates are rendered on the next animation tick.
func (s *Spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	s.message = msg
	s.mu.Unlock()
}

// StopSuccess halts the spinner and prints an [OK] status line.
func (s *Spinner) StopSuccess(msg string) {
	s.stopAnimation()
	duration := time.Since(s.startTime).Round(100 * time.Millisecond)
	pterm.Success.Printf("%s (%s)\n", msg, pterm.Gray(fmt.Sprintf("%.1fs", duration.Seconds())))
}

// StopFail halts the spinner and prints a [FAIL] status line.
func (s *Spinner) StopFail(msg string) {
	s.stopAnimation()
	pterm.Error.Println(msg)
}

// StopWarn halts the spinner and prints a [WARN] status line.
func (s *Spinner) StopWarn(msg string) {
	s.stopAnimation()
	pterm.Warning.Println(msg)
}

// StopInfo halts the spinner and prints an [*] status line.
func (s *Spinner) StopInfo(msg string) {
	s.stopAnimation()
	pterm.Info.Println(msg)
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

// RenderProgressBar returns a compact ASCII progress bar string without emojis.
func RenderProgressBar(current, total int64, width int) string {
	if total <= 0 {
		return ""
	}
	if current > total {
		current = total
	}
	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	if empty < 0 {
		empty = 0
	}

	return fmt.Sprintf("[%s%s] %3.0f%%",
		strings.Repeat("=", filled),
		strings.Repeat("-", empty),
		percent*100,
	)
}

// NewProgressBar initializes a pterm progress bar with elapsed timer and count.
func NewProgressBar(total int, title string) (*pterm.ProgressbarPrinter, error) {
	return pterm.DefaultProgressbar.
		WithTotal(total).
		WithTitle(title).
		WithShowElapsedTime(true).
		WithShowCount(true).
		WithBarCharacter("=").
		WithBarFiller("-").
		Start()
}
