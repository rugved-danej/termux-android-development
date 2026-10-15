package ui

import (
	"strings"
	"sync"

	"github.com/pterm/pterm"
)

var (
	bannerPrinted bool
	bannerMu      sync.Mutex
)

// PrintBanner outputs the full-width cyan header and author attribution once per session.
// Designed to match the look, feel, and responsiveness of termux-best-mirror.
func PrintBanner() {
	bannerMu.Lock()
	defer bannerMu.Unlock()

	if bannerPrinted {
		return
	}
	bannerPrinted = true

	pterm.DefaultCenter.Println(pterm.Cyan(pterm.DefaultHeader.WithFullWidth().Sprint("Termux Android Development")))
	pterm.Println()
}

// ResetBannerFlag resets the bannerPrinted flag.
func ResetBannerFlag() {
	bannerMu.Lock()
	defer bannerMu.Unlock()
	bannerPrinted = false
}

// MarkBannerPrinted marks the banner as displayed so sub-commands don't duplicate it.
func MarkBannerPrinted() {
	bannerMu.Lock()
	defer bannerMu.Unlock()
	bannerPrinted = true
}

// PrintBox renders a structured, responsive box around lines using pterm.
func PrintBox(title string, lines []string) {
	content := strings.Join(lines, "\n")
	pterm.DefaultBox.
		WithTitle(title).
		WithBoxStyle(pterm.NewStyle(pterm.FgCyan)).
		WithTitleTopLeft().
		Println(content)
}
