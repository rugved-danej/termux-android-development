package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

type CleanFlags struct {
	All         bool
	Gradle      bool
	SdkCache    bool
	TmpOnly     bool
	AutoConfirm bool
}

type cleanTarget struct {
	ID          string
	Label       string
	Description string
	Paths       []string
	Size        int64
	IsCommand   bool
	Command     []string
}

// RunCleanCommand scans and cleans temporary caches to reclaim device disk space.
func RunCleanCommand(args []string) error {
	ui.PrintBanner()

	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	flags := CleanFlags{}
	for _, a := range args {
		switch a {
		case "--all", "-a", "all":
			flags.All = true
		case "--gradle", "gradle":
			flags.Gradle = true
		case "--sdk", "sdk", "--sdk-cache":
			flags.SdkCache = true
		case "--tmp", "tmp":
			flags.TmpOnly = true
		case "-y", "--yes":
			flags.AutoConfirm = true
		}
	}

	targets := []cleanTarget{
		{
			ID:          "gradle-daemons",
			Label:       "Stop Gradle Daemons",
			Description: "Terminate idle background Gradle daemons to free system RAM",
			IsCommand:   true,
			Command:     []string{"gradle", "--stop"},
		},
		{
			ID:          "gradle-caches",
			Label:       "Gradle Cache & Daemons Log",
			Description: "Clear ~/.gradle/caches and ~/.gradle/daemon build logs",
			Paths: []string{
				filepath.Join(cfg.GradleDir, "caches"),
				filepath.Join(cfg.GradleDir, "daemon"),
			},
		},
		{
			ID:          "sdk-cache",
			Label:       "Android SDK Download Caches",
			Description: "Clear temporary download intermediates and ~/.android/cache",
			Paths: []string{
				filepath.Join(cfg.AndroidDir, "cache"),
				filepath.Join(cfg.SdkDir, ".downloadIntermediates"),
				filepath.Join(cfg.SdkDir, ".temp"),
			},
		},
		{
			ID:          "temp-files",
			Label:       "Temporary Update & Tool Files",
			Description: "Remove leftover binary updates and update timestamps",
			Paths: []string{
				filepath.Join(cfg.PrefixBin, "tadev.tmp"),
				filepath.Join(cfg.HomeDir, ".cache", "tadev"),
			},
		},
	}

	// Calculate disk sizes
	for i := range targets {
		var sz int64
		for _, p := range targets[i].Paths {
			sz += calculateDirSize(p)
		}
		targets[i].Size = sz
	}

	selected := make(map[string]bool)

	if flags.All {
		for _, t := range targets {
			selected[t.ID] = true
		}
	} else if flags.Gradle {
		selected["gradle-daemons"] = true
		selected["gradle-caches"] = true
	} else if flags.SdkCache {
		selected["sdk-cache"] = true
	} else if flags.TmpOnly {
		selected["temp-files"] = true
	} else if ui.IsTerminal() {
		var options []*ui.CheckboxOption
		for _, t := range targets {
			desc := t.Description
			if t.Size > 0 {
				desc = fmt.Sprintf("%s (%s)", desc, formatBytes(t.Size))
			}
			options = append(options, &ui.CheckboxOption{
				ID:          t.ID,
				Label:       t.Label,
				Description: desc,
				Checked:     true, // Default check all
			})
		}

		res, confirmed, err := ui.RunCheckboxMenu("CACHE & STORAGE CLEANER", "Select cache targets to purge from your device:", options)
		if err != nil || !confirmed {
			ui.Info("Cleanup cancelled.")
			return nil
		}
		for _, opt := range res {
			if opt.Checked {
				selected[opt.ID] = true
			}
		}
	} else {
		// Non-interactive without flags: Clean all non-destructive temporary caches
		for _, t := range targets {
			selected[t.ID] = true
		}
	}

	activeCount := 0
	for _, v := range selected {
		if v {
			activeCount++
		}
	}

	if activeCount == 0 {
		ui.Info("No cache targets selected. Nothing to clean.")
		return nil
	}

	if !flags.AutoConfirm && !ui.IsTerminal() {
		fmt.Printf("Purge selected caches? [y/N]: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			resp := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if resp != "y" && resp != "yes" {
				ui.Info("Cleanup aborted.")
				return nil
			}
		}
	}

	var totalReclaimed int64

	for _, t := range targets {
		if !selected[t.ID] {
			continue
		}

		if t.IsCommand {
			sp := ui.NewSpinner(fmt.Sprintf("%s...", t.Label))
			sp.Start()
			_ = exec.Command(t.Command[0], t.Command[1:]...).Run()
			sp.StopSuccess(t.Label + " completed.")
			continue
		}

		sp := ui.NewSpinner(fmt.Sprintf("Purging %s...", t.Label))
		sp.Start()

		for _, p := range t.Paths {
			if _, err := os.Stat(p); err == nil {
				_ = os.RemoveAll(p)
			}
		}
		totalReclaimed += t.Size
		sp.StopSuccess(fmt.Sprintf("%s purged (%s reclaimed).", t.Label, formatBytes(t.Size)))
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Cleanup complete! Reclaimed %s of disk space.", formatBytes(totalReclaimed)))
	return nil
}

func calculateDirSize(path string) int64 {
	var total int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
