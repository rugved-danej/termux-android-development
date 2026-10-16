// Package ui provides terminal styling, purple-themed ANSI color management,
// responsive tables and cards, braille spinners, and clean non-bold developer formatting
// designed to match the vtermux user interface standard.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	// ColorReset clears active terminal color styling.
	ColorReset = "\033[0m"

	// ColorPurple represents standard magenta/purple (\033[35m).
	ColorPurple = "\033[35m"

	// ColorBrightPurple represents high-intensity bright purple (\033[95m).
	ColorBrightPurple = "\033[95m"

	// ColorCyan aliases to purple to maintain a unified color palette.
	ColorCyan = "\033[35m"

	// ColorBrightCyan aliases to bright purple.
	ColorBrightCyan = "\033[95m"

	// ColorGreen represents standard terminal green (\033[32m).
	ColorGreen = "\033[32m"

	// ColorYellow represents standard terminal yellow/amber (\033[33m).
	ColorYellow = "\033[33m"

	// ColorRed represents standard terminal red (\033[31m).
	ColorRed = "\033[31m"

	// ColorBrightRed represents bright red (\033[91m).
	ColorBrightRed = "\033[91m"

	// ColorGray represents subtle muted gray (\033[90m).
	ColorGray = "\033[90m"

	// Universal terminal symbols supported across all Android and Linux terminals without emojis.
	SymbolCheck    = "✓"
	SymbolBullet   = "•"
	SymbolWarn     = "!"
	SymbolCross    = "✕"
	SymbolQuestion = "?"
	SymbolArrow    = "❯"
	SymbolSection  = "◆"

	// Backwards compatibility constants for code expecting ANSI strings directly.
	// Bold, Dim, Italic, Underline are intentionally EMPTY to prevent bold and styling.
	Reset     = "\033[0m"
	Bold      = ""
	Dim       = ""
	Italic    = ""
	Underline = ""

	Green    = "\033[32m"
	Cyan     = "\033[35m"
	Blue     = "\033[35m"
	Yellow   = "\033[33m"
	Red      = "\033[31m"
	Magenta  = "\033[35m"
	Gray     = "\033[90m"
	White    = "\033[0m"
	BgCyan   = ""
	BgGreen  = ""
	BgRed    = ""
	BgYellow = ""
)

func init() {
	// Disable all ANSI color sequences if NO_COLOR is set or TERM is dumb.
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		ColorReset = ""
		ColorPurple = ""
		ColorBrightPurple = ""
		ColorCyan = ""
		ColorBrightCyan = ""
		ColorGreen = ""
		ColorYellow = ""
		ColorRed = ""
		ColorBrightRed = ""
		ColorGray = ""

		Reset = ""
		Green = ""
		Cyan = ""
		Blue = ""
		Yellow = ""
		Red = ""
		Magenta = ""
		Gray = ""
		White = ""
	}
}

// Purple wraps text in primary ANSI purple.
func Purple(text string) string {
	if ColorPurple == "" {
		return text
	}
	return ColorPurple + text + ColorReset
}

// BrightPurple wraps text in high-intensity bright purple.
func BrightPurple(text string) string {
	if ColorBrightPurple == "" {
		return text
	}
	return ColorBrightPurple + text + ColorReset
}

// CyanText wraps text in primary purple (mapped for palette consistency).
func CyanText(text string) string {
	return Purple(text)
}

// GreenText wraps text in standard green.
func GreenText(text string) string {
	if ColorGreen == "" {
		return text
	}
	return ColorGreen + text + ColorReset
}

// YellowText wraps text in standard yellow.
func YellowText(text string) string {
	if ColorYellow == "" {
		return text
	}
	return ColorYellow + text + ColorReset
}

// RedText wraps text in standard red.
func RedText(text string) string {
	if ColorRed == "" {
		return text
	}
	return ColorRed + text + ColorReset
}

// GrayText wraps text in muted gray.
func GrayText(text string) string {
	if ColorGray == "" {
		return text
	}
	return ColorGray + text + ColorReset
}

// Colorize wraps text with specified ANSI color sequence and resets styling.
func Colorize(color, text string) string {
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}

// Success prints a green checkmark icon followed by a success message.
func Success(text string) {
	fmt.Printf("%s %s\n", GreenText(SymbolCheck), text)
}

// Successf prints a green checkmark icon followed by a formatted success message.
func Successf(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", GreenText(SymbolCheck), fmt.Sprintf(format, a...))
}

// Info prints a purple bullet icon followed by an informational message.
func Info(text string) {
	fmt.Printf("%s %s\n", Purple(SymbolBullet), text)
}

// Infof prints a purple bullet icon followed by a formatted informational message.
func Infof(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", Purple(SymbolBullet), fmt.Sprintf(format, a...))
}

// Warn prints a yellow warning icon followed by a warning message.
func Warn(text string) {
	fmt.Printf("%s %s\n", YellowText(SymbolWarn), text)
}

// Warnf prints a yellow warning icon followed by a formatted warning message.
func Warnf(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", YellowText(SymbolWarn), fmt.Sprintf(format, a...))
}

// Fail prints a red cross icon followed by an error message.
func Fail(text string) {
	fmt.Printf("%s %s\n", RedText(SymbolCross), text)
}

// Failf prints a red cross icon followed by a formatted error message.
func Failf(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", RedText(SymbolCross), fmt.Sprintf(format, a...))
}

// Error is an alias for Fail.
func Error(text string) {
	Fail(text)
}

// Errorf is an alias for Failf.
func Errorf(format string, a ...interface{}) {
	Failf(format, a...)
}

// Step prints a numbered workflow step indicator e.g. [1/6] Step title in purple.
func Step(num, total, text string) {
	fmt.Printf("%s %s\n", BrightPurple(fmt.Sprintf("[%s/%s]", num, total)), text)
}

// Confirm presents an interactive yes/no question prompt to the user on standard input.
func Confirm(promptText string, defaultYes bool) bool {
	hint := "(y/N)"
	if defaultYes {
		hint = "(Y/n)"
	}
	fmt.Printf("%s %s %s ", BrightPurple(SymbolQuestion), promptText, GrayText(hint))
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" {
		return defaultYes
	}
	return input == "y" || input == "yes"
}

// Prompt asks the user for a text input line on standard input.
func Prompt(promptText string) string {
	fmt.Printf("%s %s: ", BrightPurple(SymbolQuestion), promptText)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// VisibleLen computes the displayed monospace character count, disregarding ANSI escape sequences.
func VisibleLen(s string) int {
	inEscape := false
	length := 0
	for _, r := range s {
		if r == 0x1b {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		if r == 0xfe0f || r == 0x200d {
			continue
		}
		if (r >= 0x1f000 && r <= 0x1fa00) || (r >= 0x2600 && r <= 0x27bf) {
			length += 2
		} else {
			length += 1
		}
	}
	return length
}

// wrapText splits a string into chunks not exceeding the specified character limit.
func wrapText(s string, limit int) []string {
	if limit <= 0 || VisibleLen(s) <= limit {
		return []string{s}
	}
	var res []string
	runes := []rune(s)
	for len(runes) > 0 {
		if len(runes) <= limit {
			res = append(res, string(runes))
			break
		}
		res = append(res, string(runes[:limit]))
		runes = runes[limit:]
	}
	return res
}

// winsize matches the Linux kernel struct winsize for the TIOCGWINSZ ioctl.
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// TerminalWidth retrieves the active terminal column width using the TIOCGWINSZ ioctl,
// falling back to the $COLUMNS environment variable or a safe default of 80 columns.
func TerminalWidth() int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if c, err := strconv.Atoi(cols); err == nil && c > 0 {
			return c
		}
	}
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))
	if int(retCode) >= 0 && ws.Col > 0 {
		return int(ws.Col)
	}
	return 80
}

// TerminalHeight retrieves the active terminal row height using the TIOCGWINSZ ioctl,
// falling back to the $LINES environment variable or a safe default of 24 rows.
func TerminalHeight() int {
	if lines := os.Getenv("LINES"); lines != "" {
		if l, err := strconv.Atoi(lines); err == nil && l > 0 {
			return l
		}
	}
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))
	if int(retCode) >= 0 && ws.Row > 0 {
		return int(ws.Row)
	}
	return 24
}

// Card prints a clean, boxed card representation of key-value properties.
func Card(title string, fields [][2]string) {
	termWidth := TerminalWidth()

	if termWidth < 48 {
		width := termWidth - 2
		if width < 26 {
			width = 26
		}
		titleLen := VisibleLen(title)
		rightDashes := width - titleLen - 5
		if rightDashes < 2 {
			rightDashes = 2
		}
		fmt.Printf("%s╭─ %s %s%s\n", ColorPurple, Purple(title), strings.Repeat("─", rightDashes), ColorPurple+"╮"+ColorReset)
		for _, f := range fields {
			keyPadding := width - len(f[0]) - 5
			if keyPadding < 0 {
				keyPadding = 0
			}
			fmt.Printf("%s│%s  %s:%s  %s│%s\n", ColorPurple, ColorReset, GrayText(f[0]), strings.Repeat(" ", keyPadding), ColorPurple, ColorReset)
			maxValWidth := width - 6
			if maxValWidth < 10 {
				maxValWidth = 10
			}
			valLines := wrapText(f[1], maxValWidth)
			for _, vl := range valLines {
				valPadding := width - VisibleLen(vl) - 6
				if valPadding < 0 {
					valPadding = 0
				}
				fmt.Printf("%s│%s    %s%s  %s│%s\n", ColorPurple, ColorReset, vl, strings.Repeat(" ", valPadding), ColorPurple, ColorReset)
			}
		}
		fmt.Printf("%s╰%s╯%s\n", ColorPurple, strings.Repeat("─", width-2), ColorReset)
		return
	}

	maxFieldLen := 0
	for _, f := range fields {
		fl := len(f[0]) + VisibleLen(f[1])
		if fl > maxFieldLen {
			maxFieldLen = fl
		}
	}

	contentWidth := maxFieldLen + 8
	titleLen := VisibleLen(title)
	if contentWidth < titleLen+8 {
		contentWidth = titleLen + 8
	}

	if contentWidth > termWidth-2 {
		contentWidth = termWidth - 2
	}

	rightDashes := contentWidth - titleLen - 5
	if rightDashes < 2 {
		rightDashes = 2
	}

	fmt.Printf("%s╭─ %s %s%s\n", ColorPurple, Purple(title), strings.Repeat("─", rightDashes), ColorPurple+"╮"+ColorReset)
	for _, f := range fields {
		padding := contentWidth - len(f[0]) - VisibleLen(f[1]) - 8
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("%s│%s  %s : %s%s  %s│%s\n",
			ColorPurple, ColorReset,
			GrayText(f[0]),
			f[1],
			strings.Repeat(" ", padding),
			ColorPurple, ColorReset,
		)
	}
	fmt.Printf("%s╰%s╯%s\n", ColorPurple, strings.Repeat("─", contentWidth-2), ColorReset)
}

// PropertyTable prints an unbroken, connected two-column table with box-drawing characters.
func PropertyTable(fields [][2]string) {
	termWidth := TerminalWidth()

	if termWidth < 50 {
		width := termWidth - 2
		if width < 30 {
			width = 30
		}
		top := ColorPurple + "┌" + strings.Repeat("─", width-2) + "┐" + ColorReset
		mid := ColorPurple + "├" + strings.Repeat("─", width-2) + "┤" + ColorReset
		bot := ColorPurple + "└" + strings.Repeat("─", width-2) + "┘" + ColorReset

		fmt.Println(top)
		for i, f := range fields {
			keyPadding := width - len(f[0]) - 4
			if keyPadding < 0 {
				keyPadding = 0
			}
			fmt.Printf("%s│%s  %s%s%s│%s\n", ColorPurple, ColorReset, Purple(f[0]), strings.Repeat(" ", keyPadding), ColorPurple, ColorReset)

			valLines := wrapText(f[1], width-6)
			for _, vl := range valLines {
				valPadding := width - VisibleLen(vl) - 6
				if valPadding < 0 {
					valPadding = 0
				}
				fmt.Printf("%s│%s    %s%s%s│%s\n", ColorPurple, ColorReset, vl, strings.Repeat(" ", valPadding), ColorPurple, ColorReset)
			}
			if i < len(fields)-1 {
				fmt.Println(mid)
			}
		}
		fmt.Println(bot)
		return
	}

	keyWidth := len("PROPERTY")
	for _, f := range fields {
		if len(f[0]) > keyWidth {
			keyWidth = len(f[0])
		}
	}
	keyColWidth := keyWidth + 2

	tableWidth := termWidth - 2
	if tableWidth > 78 {
		tableWidth = 78
	}

	valColWidth := tableWidth - keyColWidth - 5
	if valColWidth < 20 {
		valColWidth = 20
		tableWidth = keyColWidth + valColWidth + 5
	}

	top := ColorPurple + "┌" + strings.Repeat("─", keyColWidth+2) + "┬" + strings.Repeat("─", valColWidth+2) + "┐" + ColorReset
	mid := ColorPurple + "├" + strings.Repeat("─", keyColWidth+2) + "┼" + strings.Repeat("─", valColWidth+2) + "┤" + ColorReset
	bot := ColorPurple + "└" + strings.Repeat("─", keyColWidth+2) + "┴" + strings.Repeat("─", valColWidth+2) + "┘" + ColorReset

	fmt.Println(top)
	keyHdrPadding := keyColWidth - len("PROPERTY")
	valHdrPadding := valColWidth - len("VALUE")
	fmt.Printf("%s│%s %s%s %s│%s %s%s %s│%s\n",
		ColorPurple, ColorReset,
		Purple("PROPERTY"), strings.Repeat(" ", keyHdrPadding),
		ColorPurple, ColorReset,
		Purple("VALUE"), strings.Repeat(" ", valHdrPadding),
		ColorPurple, ColorReset,
	)
	fmt.Println(mid)

	for _, f := range fields {
		valLines := wrapText(f[1], valColWidth)
		for lineIdx, vl := range valLines {
			keyText := ""
			if lineIdx == 0 {
				keyText = f[0]
			}
			kPad := keyColWidth - VisibleLen(keyText)
			if kPad < 0 {
				kPad = 0
			}
			vPad := valColWidth - VisibleLen(vl)
			if vPad < 0 {
				vPad = 0
			}
			fmt.Printf("%s│%s %s%s %s│%s %s%s %s│%s\n",
				ColorPurple, ColorReset,
				keyText, strings.Repeat(" ", kPad),
				ColorPurple, ColorReset,
				vl, strings.Repeat(" ", vPad),
				ColorPurple, ColorReset,
			)
		}
	}
	fmt.Println(bot)
}

// Table renders tabular data with clean box-drawing borders.
func Table(headers []string, rows [][]string) {
	termWidth := TerminalWidth()

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = VisibleLen(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				vl := VisibleLen(cell)
				if vl > colWidths[i] {
					colWidths[i] = vl
				}
			}
		}
	}

	totalWidth := 1
	for _, w := range colWidths {
		totalWidth += w + 3
	}

	if totalWidth > termWidth {
		for i, row := range rows {
			var title string
			if len(row) > 0 {
				title = row[0]
			}
			var fields [][2]string
			for j := 1; j < len(row) && j < len(headers); j++ {
				fields = append(fields, [2]string{headers[j], row[j]})
			}
			Card(title, fields)
			if i < len(rows)-1 {
				fmt.Println()
			}
		}
		return
	}

	var top, mid, bot strings.Builder
	top.WriteString(ColorPurple + "┌")
	mid.WriteString(ColorPurple + "├")
	bot.WriteString(ColorPurple + "└")

	for i, w := range colWidths {
		dash := strings.Repeat("─", w+2)
		top.WriteString(dash)
		mid.WriteString(dash)
		bot.WriteString(dash)

		if i < len(colWidths)-1 {
			top.WriteString("┬")
			mid.WriteString("┼")
			bot.WriteString("┴")
		} else {
			top.WriteString("┐" + ColorReset)
			mid.WriteString("┤" + ColorReset)
			bot.WriteString("┘" + ColorReset)
		}
	}

	fmt.Println(top.String())

	fmt.Print(ColorPurple + "│" + ColorReset)
	for i, h := range headers {
		padding := colWidths[i] - VisibleLen(h)
		fmt.Printf(" %s%s %s│%s", Purple(h), strings.Repeat(" ", padding), ColorPurple, ColorReset)
	}
	fmt.Println()

	fmt.Println(mid.String())

	for _, row := range rows {
		fmt.Print(ColorPurple + "│" + ColorReset)
		for i, cell := range row {
			if i < len(colWidths) {
				padding := colWidths[i] - VisibleLen(cell)
				fmt.Printf(" %s%s %s│%s", cell, strings.Repeat(" ", padding), ColorPurple, ColorReset)
			}
		}
		fmt.Println()
	}

	fmt.Println(bot.String())
}
