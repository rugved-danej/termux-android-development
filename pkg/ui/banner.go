package ui

import (
	"fmt"
	"strings"
	"sync"
)

var (
	bannerPrinted bool
	bannerMu      sync.Mutex
)

// PrintBanner outputs nothing to match vtermux clean interface style.
func PrintBanner() {
	// No-op to match clean vtermux style
}

// ResetBannerFlag resets the bannerPrinted flag.
func ResetBannerFlag() {
	bannerMu.Lock()
	defer bannerMu.Unlock()
	bannerPrinted = false
}

// MarkBannerPrinted marks the banner as displayed.
func MarkBannerPrinted() {
	bannerMu.Lock()
	defer bannerMu.Unlock()
	bannerPrinted = true
}

// PrintBox renders a structured, responsive purple box around lines matching vtermux card/box style.
func PrintBox(title string, lines []string) {
	termWidth := TerminalWidth()
	maxLen := VisibleLen(title)
	for _, l := range lines {
		vl := VisibleLen(l)
		if vl > maxLen {
			maxLen = vl
		}
	}

	width := maxLen + 6
	if width < 30 {
		width = 30
	}
	if width > termWidth-2 {
		width = termWidth - 2
	}

	rightDashes := width - VisibleLen(title) - 5
	if rightDashes < 2 {
		rightDashes = 2
	}

	fmt.Printf("%s╭─ %s %s%s\n", ColorPurple, Purple(title), strings.Repeat("─", rightDashes), ColorPurple+"╮"+ColorReset)
	for _, l := range lines {
		pad := width - VisibleLen(l) - 4
		if pad < 0 {
			pad = 0
		}
		fmt.Printf("%s│%s  %s%s%s│%s\n", ColorPurple, ColorReset, l, strings.Repeat(" ", pad), ColorPurple, ColorReset)
	}
	fmt.Printf("%s╰%s╯%s\n", ColorPurple, strings.Repeat("─", width-2), ColorReset)
}
