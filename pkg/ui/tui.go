package ui

import (
	"fmt"
	"os"
	"strings"

	"atomicgo.dev/cursor"
	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/pterm/pterm"
)

// IsTerminal checks if standard input is attached to an interactive terminal device.
func IsTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

// GetTerminalSize returns the current columns and rows of the terminal window.
func GetTerminalSize() (width int, height int) {
	w := pterm.GetTerminalWidth()
	h := pterm.GetTerminalHeight()
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	return w, h
}

// CheckboxOption represents a selectable component in the interactive TUI.
type CheckboxOption struct {
	ID          string
	Label       string
	Description string
	Checked     bool
}

// MenuItem represents an action entry in the interactive dashboard TUI.
type MenuItem struct {
	Key         string
	Title       string
	Description string
}

func visibleLen(s string) int {
	return len([]rune(stripANSI(s)))
}

func truncateLine(s string, maxCols int) string {
	if maxCols <= 0 {
		return ""
	}
	vLen := visibleLen(s)
	if vLen <= maxCols {
		return s
	}
	plain := stripANSI(s)
	runes := []rune(plain)
	if len(runes) > maxCols {
		return string(runes[:maxCols])
	}
	return s
}

func padBoxLine(line string, width int) string {
	vLen := visibleLen(line)
	contentLen := vLen - 4
	if contentLen < 0 {
		contentLen = 0
	}
	targetInside := width - 4
	pad := targetInside - contentLen
	if pad < 0 {
		pad = 0
	}
	return line + strings.Repeat(" ", pad) + " │"
}

func wrapText(text string, maxLen int) []string {
	if maxLen <= 10 {
		maxLen = 40
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	curr := words[0]
	for _, w := range words[1:] {
		if len(curr)+1+len(w) <= maxLen {
			curr += " " + w
		} else {
			lines = append(lines, curr)
			curr = w
		}
	}
	lines = append(lines, curr)
	return lines
}

// RunSelectMenu displays an interactive dashboard menu with live Ctrl+D detailed description support.
// Prevents line duplication, supports keyboard shortcuts, and responds smoothly across all terminal widths.
func RunSelectMenu(title string, deviceDesc string, items []MenuItem) (int, bool, error) {
	if !IsTerminal() || len(items) == 0 {
		return 0, false, nil
	}

	MarkBannerPrinted()

	cursorIdx := 0
	showDetails := false
	confirmed := false

	maxTitleLen := 0
	for _, it := range items {
		if len(it.Title) > maxTitleLen {
			maxTitleLen = len(it.Title)
		}
	}

	defaultTitle := "Termux Android Development"
	if title != "" {
		defaultTitle = title
	}

	render := func() string {
		cols, rows := GetTerminalSize()
		if cols <= 0 {
			cols = 80
		}
		if rows <= 0 {
			rows = 24
		}
		var b strings.Builder

		b.WriteString(truncateLine(fmt.Sprintf("%s%s%s%s", Cyan, Bold, defaultTitle, Reset), cols-2) + "\n")
		headerLines := 2
		if deviceDesc != "" {
			b.WriteString(truncateLine(fmt.Sprintf("  %s[*]%s Device: %s%s%s", Cyan+Bold, Reset, Bold, deviceDesc, Reset), cols-2) + "\n")
			headerLines = 3
		}
		b.WriteString("\n")

		boxWidth := cols - 4
		if boxWidth > 76 {
			boxWidth = 76
		}
		if boxWidth < 35 {
			boxWidth = 35
		}

		footerLines := 2
		indicatorLines := 2

		// Calculate details box height if open
		detailsHeight := 0
		var descLines []string
		if showDetails && cursorIdx >= 0 && cursorIdx < len(items) {
			maxDescLines := rows - headerLines - footerLines - indicatorLines - 4 - 3
			if maxDescLines < 1 {
				maxDescLines = 1
			}
			descLines = wrapText(items[cursorIdx].Description, boxWidth-14)
			if len(descLines) > maxDescLines {
				descLines = descLines[:maxDescLines]
			}
			detailsHeight = len(descLines) + 4
		}

		fixedLines := headerLines + footerLines + detailsHeight + indicatorLines
		maxVisible := rows - fixedLines - 1
		if maxVisible < 3 {
			maxVisible = 3
		}
		if maxVisible > len(items) {
			maxVisible = len(items)
		}

		startIdx := 0
		endIdx := len(items)
		if len(items) > maxVisible {
			startIdx = cursorIdx - maxVisible/2
			if startIdx < 0 {
				startIdx = 0
			}
			endIdx = startIdx + maxVisible
			if endIdx > len(items) {
				endIdx = len(items)
				startIdx = endIdx - maxVisible
				if startIdx < 0 {
					startIdx = 0
				}
			}
		}

		if startIdx > 0 {
			b.WriteString(truncateLine(fmt.Sprintf("  %s^ (%d more actions above)%s", Gray, startIdx, Reset), cols-2) + "\n")
		}

		for i := startIdx; i < endIdx; i++ {
			it := items[i]
			prefix := "    "
			style := White
			if i == cursorIdx {
				prefix = fmt.Sprintf("  %s> %s", Cyan+Bold, Reset)
				style = Cyan + Bold
			}

			var line string
			if !showDetails {
				if cols >= 65 && it.Description != "" {
					padding := strings.Repeat(" ", maxTitleLen-len(it.Title)+2)
					avail := cols - maxTitleLen - 12
					desc := it.Description
					if len(desc) > avail && avail > 3 {
						desc = desc[:avail-3] + "..."
					}
					line = fmt.Sprintf("%s%s%s%s- %s%s", prefix, style, it.Title, Reset+padding, Gray, desc)
				} else {
					line = fmt.Sprintf("%s%s%s%s", prefix, style, it.Title, Reset)
				}
			} else {
				line = fmt.Sprintf("%s%s%s%s", prefix, style, it.Title, Reset)
			}
			b.WriteString(truncateLine(line, cols-2) + "\n")
		}

		if endIdx < len(items) {
			b.WriteString(truncateLine(fmt.Sprintf("  %sv (%d more actions below)%s", Gray, len(items)-endIdx, Reset), cols-2) + "\n")
		}

		// Detailed description panel toggled with Ctrl+D
		if showDetails && cursorIdx >= 0 && cursorIdx < len(items) {
			curItem := items[cursorIdx]
			b.WriteString("\n")

			topBorder := fmt.Sprintf("  ┌─ %sDetailed Description%s %s┐", Cyan+Bold, Reset, strings.Repeat("─", boxWidth-26))
			b.WriteString(truncateLine(topBorder, cols-2) + "\n")

			actionLine := fmt.Sprintf("  │ %sAction:%s   %s", Cyan+Bold, Reset, curItem.Title)
			b.WriteString(truncateLine(padBoxLine(actionLine, boxWidth), cols-2) + "\n")

			for idx, dl := range descLines {
				var lineStr string
				if idx == 0 {
					lineStr = fmt.Sprintf("  │ %sDetails:%s  %s", Gray, Reset, dl)
				} else {
					lineStr = fmt.Sprintf("  │           %s", dl)
				}
				b.WriteString(truncateLine(padBoxLine(lineStr, boxWidth), cols-2) + "\n")
			}

			botBorder := fmt.Sprintf("  └%s┘", strings.Repeat("─", boxWidth-2))
			b.WriteString(truncateLine(botBorder, cols-2) + "\n")
		}

		b.WriteString("\n")
		toggleHint := "[Ctrl+D] Detailed Info"
		if showDetails {
			toggleHint = "[Ctrl+D] Hide Details"
		}
		b.WriteString(truncateLine(fmt.Sprintf("  %s[Up/Down] Navigate  [Enter] Select  %s%s%s  [Esc] Exit%s", Gray, Cyan, toggleHint, Gray, Reset), cols-2))

		return b.String()
	}

	fmt.Print("\033[?1049h\033[H\033[2J")
	cursor.Hide()
	defer func() {
		cursor.Show()
		fmt.Print("\033[?1049l")
	}()

	draw := func() {
		fmt.Print("\033[H" + render() + "\033[J")
	}

	draw()

	err := keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.Up:
			if cursorIdx > 0 {
				cursorIdx--
			} else {
				cursorIdx = len(items) - 1
			}
			draw()

		case keys.Down:
			if cursorIdx < len(items)-1 {
				cursorIdx++
			} else {
				cursorIdx = 0
			}
			draw()

		case keys.CtrlD:
			showDetails = !showDetails
			draw()

		case keys.RuneKey:
			if len(key.Runes) > 0 {
				r := key.Runes[0]
				if r == 'k' || r == 'K' {
					if cursorIdx > 0 {
						cursorIdx--
					} else {
						cursorIdx = len(items) - 1
					}
					draw()
				} else if r == 'j' || r == 'J' {
					if cursorIdx < len(items)-1 {
						cursorIdx++
					} else {
						cursorIdx = 0
					}
					draw()
				} else if r == 'd' || r == 'D' {
					showDetails = !showDetails
					draw()
				} else if r == 'q' || r == 'Q' {
					confirmed = false
					return true, nil
				} else if r >= '1' && int(r-'1') < len(items) {
					cursorIdx = int(r - '1')
					draw()
				}
			}

		case keys.Enter:
			confirmed = true
			return true, nil

		case keys.Escape, keys.CtrlC:
			confirmed = false
			return true, nil
		}
		return false, nil
	})

	if err != nil || !confirmed {
		return 0, false, nil
	}

	pterm.Success.Printf("Selected: %s\n\n", pterm.Cyan(items[cursorIdx].Title))
	return cursorIdx, true, nil
}

// RunCheckboxMenu displays an in-place interactive multiselect menu without line duplications or search prompts.
func RunCheckboxMenu(title, subtitle string, items []*CheckboxOption) ([]*CheckboxOption, bool, error) {
	if !IsTerminal() || len(items) == 0 {
		return items, true, nil
	}

	MarkBannerPrinted()

	cursorIdx := 0
	confirmed := false

	maxLabelLen := 0
	for _, it := range items {
		if len(it.Label) > maxLabelLen {
			maxLabelLen = len(it.Label)
		}
	}

	defaultTitle := "Select Components"
	if title != "" {
		defaultTitle = title
	}

	render := func() string {
		cols, rows := GetTerminalSize()
		if cols <= 0 {
			cols = 80
		}
		if rows <= 0 {
			rows = 24
		}
		var b strings.Builder

		b.WriteString(truncateLine(fmt.Sprintf("%s%s%s%s", Cyan, Bold, defaultTitle, Reset), cols-2) + "\n")
		headerLines := 2
		if subtitle != "" {
			b.WriteString(truncateLine(fmt.Sprintf("  %s[*]%s %s", Cyan+Bold, Reset, subtitle), cols-2) + "\n")
			headerLines = 3
		}
		b.WriteString("\n")

		footerLines := 2
		indicatorLines := 2
		fixedLines := headerLines + footerLines + indicatorLines
		maxVisible := rows - fixedLines - 1
		if maxVisible < 3 {
			maxVisible = 3
		}
		if maxVisible > len(items) {
			maxVisible = len(items)
		}

		startIdx := 0
		endIdx := len(items)
		if len(items) > maxVisible {
			startIdx = cursorIdx - maxVisible/2
			if startIdx < 0 {
				startIdx = 0
			}
			endIdx = startIdx + maxVisible
			if endIdx > len(items) {
				endIdx = len(items)
				startIdx = endIdx - maxVisible
				if startIdx < 0 {
					startIdx = 0
				}
			}
		}

		if startIdx > 0 {
			b.WriteString(truncateLine(fmt.Sprintf("  %s^ (%d more components above)%s", Gray, startIdx, Reset), cols-2) + "\n")
		}

		for i := startIdx; i < endIdx; i++ {
			it := items[i]
			pointer := "  "
			style := White
			if i == cursorIdx {
				pointer = fmt.Sprintf("%s> %s", Cyan+Bold, Reset)
				style = Cyan + Bold
			}

			checkStr := "[ ] "
			if it.Checked {
				checkStr = fmt.Sprintf("[%sx%s] ", Green+Bold, Reset)
			}

			var line string
			if cols >= 65 && it.Description != "" {
				padding := strings.Repeat(" ", maxLabelLen-len(it.Label)+2)
				avail := cols - maxLabelLen - 14
				desc := it.Description
				if len(desc) > avail && avail > 3 {
					desc = desc[:avail-3] + "..."
				}
				line = fmt.Sprintf("  %s%s%s%s%s- %s%s", pointer, checkStr, style, it.Label, Reset+padding, Gray, desc)
			} else {
				line = fmt.Sprintf("  %s%s%s%s%s", pointer, checkStr, style, it.Label, Reset)
			}
			b.WriteString(truncateLine(line, cols-2) + "\n")
		}

		if endIdx < len(items) {
			b.WriteString(truncateLine(fmt.Sprintf("  %sv (%d more components below)%s", Gray, len(items)-endIdx, Reset), cols-2) + "\n")
		}

		b.WriteString("\n")
		b.WriteString(truncateLine(fmt.Sprintf("  %s[Up/Down] Navigate  [Space] Toggle  [a] Toggle All  [Enter] Confirm  [Esc] Cancel%s", Gray, Reset), cols-2))

		return b.String()
	}

	fmt.Print("\033[?1049h\033[H\033[2J")
	cursor.Hide()
	defer func() {
		cursor.Show()
		fmt.Print("\033[?1049l")
	}()

	draw := func() {
		fmt.Print("\033[H" + render() + "\033[J")
	}

	draw()

	err := keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.Up:
			if cursorIdx > 0 {
				cursorIdx--
			} else {
				cursorIdx = len(items) - 1
			}
			draw()

		case keys.Down:
			if cursorIdx < len(items)-1 {
				cursorIdx++
			} else {
				cursorIdx = 0
			}
			draw()

		case keys.Space:
			items[cursorIdx].Checked = !items[cursorIdx].Checked
			draw()

		case keys.RuneKey:
			if len(key.Runes) > 0 {
				r := key.Runes[0]
				if r == 'k' || r == 'K' {
					if cursorIdx > 0 {
						cursorIdx--
					} else {
						cursorIdx = len(items) - 1
					}
					draw()
				} else if r == 'j' || r == 'J' {
					if cursorIdx < len(items)-1 {
						cursorIdx++
					} else {
						cursorIdx = 0
					}
					draw()
				} else if r == 'a' || r == 'A' {
					allChecked := true
					for _, it := range items {
						if !it.Checked {
							allChecked = false
							break
						}
					}
					for _, it := range items {
						it.Checked = !allChecked
					}
					draw()
				} else if r == 'q' || r == 'Q' {
					confirmed = false
					return true, nil
				}
			}

		case keys.Enter:
			confirmed = true
			return true, nil

		case keys.Escape, keys.CtrlC:
			confirmed = false
			return true, nil
		}
		return false, nil
	})

	if err != nil || !confirmed {
		return items, false, nil
	}

	selectedCount := 0
	for _, it := range items {
		if it.Checked {
			selectedCount++
		}
	}

	pterm.Success.Printf("Selected %d component(s) for uninstallation.\n\n", selectedCount)
	return items, true, nil
}

// RunJavaProfileMenu displays an interactive Java profile selector without inline descriptions.
// Users navigate with Up/Down, select with Enter, or press Tab to open the GitHub documentation
// with detailed comparisons between combinations and solos directly in their default browser.
func RunJavaProfileMenu(title string, deviceDesc string, items []MenuItem, helpURL string) (int, bool, error) {
	if !IsTerminal() || len(items) == 0 {
		return 0, false, nil
	}

	MarkBannerPrinted()

	cursorIdx := 0
	confirmed := false
	statusMsg := ""

	defaultTitle := "SELECT JAVA PROFILE"
	if title != "" {
		defaultTitle = title
	}

	render := func() string {
		cols, rows := GetTerminalSize()
		if cols <= 0 {
			cols = 80
		}
		if rows <= 0 {
			rows = 24
		}
		var b strings.Builder

		b.WriteString(truncateLine(fmt.Sprintf("%s%s%s%s", Cyan, Bold, defaultTitle, Reset), cols-2) + "\n")
		headerLines := 2
		if deviceDesc != "" {
			b.WriteString(truncateLine(fmt.Sprintf("  %s[*]%s Device: %s%s%s", Cyan+Bold, Reset, Bold, deviceDesc, Reset), cols-2) + "\n")
			headerLines = 3
		}
		b.WriteString("\n")

		statusLines := 0
		if statusMsg != "" {
			statusLines = 1
		}

		footerLines := 2
		indicatorLines := 2
		fixedLines := headerLines + footerLines + statusLines + indicatorLines
		maxVisible := rows - fixedLines - 1
		if maxVisible < 3 {
			maxVisible = 3
		}
		if maxVisible > len(items) {
			maxVisible = len(items)
		}

		startIdx := 0
		endIdx := len(items)
		if len(items) > maxVisible {
			startIdx = cursorIdx - maxVisible/2
			if startIdx < 0 {
				startIdx = 0
			}
			endIdx = startIdx + maxVisible
			if endIdx > len(items) {
				endIdx = len(items)
				startIdx = endIdx - maxVisible
				if startIdx < 0 {
					startIdx = 0
				}
			}
		}

		if startIdx > 0 {
			b.WriteString(truncateLine(fmt.Sprintf("  %s^ (%d more options above)%s", Gray, startIdx, Reset), cols-2) + "\n")
		}

		for i := startIdx; i < endIdx; i++ {
			it := items[i]
			prefix := "    "
			style := White
			if i == cursorIdx {
				prefix = fmt.Sprintf("  %s> %s", Cyan+Bold, Reset)
				style = Cyan + Bold
			}
			line := fmt.Sprintf("%s%s%s%s", prefix, style, it.Title, Reset)
			b.WriteString(truncateLine(line, cols-2) + "\n")
		}

		if endIdx < len(items) {
			b.WriteString(truncateLine(fmt.Sprintf("  %sv (%d more options below)%s", Gray, len(items)-endIdx, Reset), cols-2) + "\n")
		}

		b.WriteString("\n")
		if statusMsg != "" {
			b.WriteString(truncateLine(fmt.Sprintf("  %s%s%s\n", Green, statusMsg, Reset), cols-2))
		}
		b.WriteString(truncateLine(fmt.Sprintf("  %s[Up/Down] Navigate  [Enter] Select  %s[Tab] Differences (Docs)%s  [Esc] Exit%s", Gray, Cyan, Gray, Reset), cols-2))

		return b.String()
	}

	fmt.Print("\033[?1049h\033[H\033[2J")
	cursor.Hide()
	defer func() {
		cursor.Show()
		fmt.Print("\033[?1049l")
	}()

	draw := func() {
		fmt.Print("\033[H" + render() + "\033[J")
	}

	draw()

	err := keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.Up:
			if cursorIdx > 0 {
				cursorIdx--
			} else {
				cursorIdx = len(items) - 1
			}
			draw()

		case keys.Down:
			if cursorIdx < len(items)-1 {
				cursorIdx++
			} else {
				cursorIdx = 0
			}
			draw()

		case keys.Tab:
			if helpURL != "" {
				go func() {
					_ = OpenBrowser(helpURL)
				}()
				statusMsg = "[OK] Opening comparison guide in browser..."
				draw()
			}

		case keys.RuneKey:
			if len(key.Runes) > 0 {
				r := key.Runes[0]
				if r == '\t' {
					if helpURL != "" {
						go func() {
							_ = OpenBrowser(helpURL)
						}()
						statusMsg = "[OK] Opening comparison guide in browser..."
						draw()
					}
				} else if r == 'k' || r == 'K' {
					if cursorIdx > 0 {
						cursorIdx--
					} else {
						cursorIdx = len(items) - 1
					}
					draw()
				} else if r == 'j' || r == 'J' {
					if cursorIdx < len(items)-1 {
						cursorIdx++
					} else {
						cursorIdx = 0
					}
					draw()
				} else if r == 'q' || r == 'Q' {
					confirmed = false
					return true, nil
				} else if r >= '1' && int(r-'1') < len(items) {
					cursorIdx = int(r - '1')
					draw()
				}
			}

		case keys.Enter:
			confirmed = true
			return true, nil

		case keys.Escape, keys.CtrlC:
			confirmed = false
			return true, nil
		}
		return false, nil
	})

	if err != nil || !confirmed {
		return 0, false, nil
	}
	return cursorIdx, true, nil
}
