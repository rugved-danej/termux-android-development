package ui

import (
	"fmt"
	"os/exec"
)

// OpenBrowser opens targetURL in the user's default browser across Termux Android and Linux.
// It relies strictly on standard built-in Termux tools ('termux-open-url', 'am start --user 0')
// and Linux desktop openers ('xdg-open', 'sensible-browser').
// It has ZERO dependency on the Termux:API companion app or the termux-api package.
func OpenBrowser(targetURL string) error {
	candidates := []struct {
		cmd  string
		args []string
	}{
		// 1. Termux default open url helper (from termux-tools, standard base package)
		{"termux-open-url", []string{targetURL}},
		// 2. Direct Termux am wrapper (standard base package, zero termux-api required)
		{"am", []string{"start", "--user", "0", "-a", "android.intent.action.VIEW", "-d", targetURL}},
		// 3. Termux-open generic launcher
		{"termux-open", []string{targetURL}},
		// 4. Linux standard XDG desktop opener
		{"xdg-open", []string{targetURL}},
		// 5. Debian/Ubuntu sensible-browser
		{"sensible-browser", []string{targetURL}},
		// 6. GNOME gio
		{"gio", []string{"open", targetURL}},
		// 7. Python standard library webbrowser module fallback
		{"python3", []string{"-m", "webbrowser", targetURL}},
	}

	for _, c := range candidates {
		if path, err := exec.LookPath(c.cmd); err == nil {
			cmd := exec.Command(path, c.args...)
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Start(); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("no supported browser opener found (tried termux-open-url, am, xdg-open)")
}
