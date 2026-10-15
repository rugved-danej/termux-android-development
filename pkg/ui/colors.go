package ui

import (
	"fmt"

	"github.com/pterm/pterm"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	// Colors
	Green   = "\033[38;5;82m"
	Cyan    = "\033[38;5;87m"
	Blue    = "\033[38;5;75m"
	Yellow  = "\033[38;5;220m"
	Red     = "\033[38;5;196m"
	Magenta = "\033[38;5;213m"
	Gray    = "\033[38;5;244m"
	White   = "\033[38;5;255m"

	// Backgrounds
	BgCyan   = "\033[48;5;31m"
	BgGreen  = "\033[48;5;28m"
	BgRed    = "\033[48;5;124m"
	BgYellow = "\033[48;5;136m"
)

func init() {
	pterm.EnableColor()

	// Zero-emoji clean developer prefixes (matching termux-best-mirror style)
	pterm.Success = *pterm.Success.WithPrefix(pterm.Prefix{
		Text:  "[OK]",
		Style: pterm.NewStyle(pterm.FgGreen, pterm.Bold),
	})
	pterm.Error = *pterm.Error.WithPrefix(pterm.Prefix{
		Text:  "[FAIL]",
		Style: pterm.NewStyle(pterm.FgRed, pterm.Bold),
	})
	pterm.Warning = *pterm.Warning.WithPrefix(pterm.Prefix{
		Text:  "[WARN]",
		Style: pterm.NewStyle(pterm.FgYellow, pterm.Bold),
	})
	pterm.Info = *pterm.Info.WithPrefix(pterm.Prefix{
		Text:  "[*]",
		Style: pterm.NewStyle(pterm.FgCyan, pterm.Bold),
	})
}

func Colorize(color, text string) string {
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}

func Success(text string) {
	pterm.Success.Println(text)
}

func Fail(text string) {
	pterm.Error.Println(text)
}

func Warn(text string) {
	pterm.Warning.Println(text)
}

func Info(text string) {
	pterm.Info.Println(text)
}

func Step(num, total, text string) {
	fmt.Printf("\n%s[%s/%s]%s %s%s%s\n", Cyan+Bold, num, total, Reset, Bold, text, Reset)
}
