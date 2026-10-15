// Package main is the entrypoint for the tadev CLI tool.
package main

import (
	"fmt"
	"os"
	"strings"

	"tadev/cmd"
	"tadev/pkg/env"
	"tadev/pkg/ui"
	"tadev/pkg/updater"
)

var (
	Version = "1.0.0"
	Author  = "Rugved Danej (@rugved-danej)"
	Repo    = "https://github.com/rugved-danej/termux-android-development"
)

func showHelp() {
	ui.PrintBanner()
	fmt.Printf(`%sUSAGE:%s
  tadev <command> [options]

%sCORE ENVIRONMENT COMMANDS:%s
  %sinstall%s, setup          Install complete Android development environment
  %sjava%s [switch|list|test] Manage OpenJDK profiles (21, 25+21, 25+17, 21+17, 25, 17)
  %suninstall%s, remove        Granular uninstallation (interactive chooser or flags)
  %supdate%s, upgrade          Download and apply all available updates (SDK, JDK, tools)
  %scheck%s, check-updates     Check for updates (SDK tools, package repos, and tadev CLI)
  %ssdk%s [list|install <pkg>] Inspect installed platforms or install extra SDK packages
  %sndk%s [status|install|list] Manage Android NDK native toolchains & shebangs
  %sclean%s [all|gradle|sdk]   Purge temporary build caches, daemon logs, and reclaim storage
  %scompletion%s [shell]       Generate or install native shell tab-completion (bash/zsh/fish)
  %sdoctor%s, status           Run diagnostic health checks on all components
  %sconfig%s [action]          Inspect or initialize configuration (show/init/reset/path)
  %sself-update%s              Upgrade the tadev CLI binary directly from GitHub
  %smenu%s, tui, dashboard     Launch interactive terminal UI dashboard
  %sversion%s, -v              Show tadev version & author information
  %shelp%s, -h                 Display this guide

%sJAVA PROFILES:%s
  21                       OpenJDK 21 LTS Only (Single JDK, 100%% AGP Stable, Saves ~300MB) [Recommended]
  25+21                    OpenJDK 25 (Host) + OpenJDK 21 (Gradle Toolchain)
  25+17                    OpenJDK 25 (Host) + OpenJDK 17 (Gradle Toolchain)
  21+17                    OpenJDK 21 (Host) + OpenJDK 17 (Gradle Toolchain)
  25                       OpenJDK 25 LTS Only (Single JDK, Modern 2025 LTS Standard)
  17                       OpenJDK 17 LTS Only (Single JDK, AGP 8.0 Baseline)

%sUNINSTALL OPTIONS:%s
  --all                    Purge EVERYTHING: SDK, JDKs, packages, config, and self-remove
  --self                   Self-remove tadev and termux-android-dev binaries from system
  --packages               Purge installed development packages (Java 25, 21, 17, Gradle, AAPT2)
  --sdk-only               Delete only the Android SDK directory (~/.android/sdk)
  -y, --yes                Skip confirmation prompts

%sGENERAL OPTIONS:%s
  --java <profile>         Specify Java environment profile (e.g. 21, 25+21, 25+17, 21+17, 25, 17)
  --force                  Force re-installation / re-download of tools
  --quiet                  Quiet output mode (for background cron/shell checks)

%sEXAMPLES:%s
  tadev                             # Launch interactive TUI Dashboard
  tadev install                     # Interactive setup (prompts for Java profile)
  tadev install --java 21           # Headless installation with single OpenJDK 21 LTS
  tadev java switch 21              # Switch active profile to single OpenJDK 21 LTS
  tadev java test                   # Test all installed Java versions (javac, gradle, d8, apksigner)
  tadev ndk install                 # Install recommended LTS Android NDK toolchains
  tadev clean                       # Reclaim device storage by purging temporary build caches
  tadev completion install          # Install native tab-completion for active shell
  tadev doctor                      # Diagnose system health & architecture
  tadev uninstall                   # Interactive uninstallation chooser
`,
		ui.Bold+ui.Yellow, ui.Reset,
		ui.Bold+ui.Cyan, ui.Reset,
		ui.Bold+ui.Green, ui.Reset,
		ui.Bold+ui.Cyan, ui.Reset,
		ui.Bold+ui.Red, ui.Reset,
		ui.Bold+ui.Green, ui.Reset,
		ui.Bold+ui.Cyan, ui.Reset,
		ui.Bold+ui.Magenta, ui.Reset,
		ui.Bold+ui.Yellow, ui.Reset,
		ui.Bold+ui.Green, ui.Reset,
		ui.Bold+ui.Cyan, ui.Reset,
		ui.Bold+ui.Blue, ui.Reset,
		ui.Bold+ui.Yellow, ui.Reset,
		ui.Bold+ui.Gray, ui.Reset,
		ui.Bold+ui.Cyan, ui.Reset,
		ui.Bold+ui.Gray, ui.Reset,
		ui.Bold+ui.Cyan, ui.Reset,
		ui.Bold+ui.Cyan, ui.Reset,
		ui.Bold+ui.Yellow, ui.Reset,
		ui.Bold+ui.Yellow, ui.Reset,
		ui.Bold+ui.Yellow, ui.Reset,
	)
}

func runDashboard(cfg *env.Config) {
	menuItems := []ui.MenuItem{
		{Key: "1", Title: "Install / Setup Environment", Description: "Configure OpenJDK, Android SDK, AAPT2, and shell paths"},
		{Key: "2", Title: "Java Profile Manager", Description: "Switch or test active OpenJDK profiles (21, 25+21, 25+17, 21+17, 25, 17)"},
		{Key: "3", Title: "Apply All Available Updates", Description: "Download and apply all upgrades for SDK, tools, and packages"},
		{Key: "4", Title: "Check for Updates", Description: "Check Google repositories, package repos, and GitHub releases"},
		{Key: "5", Title: "SDK Package Manager", Description: "List installed platforms or install extra SDK packages"},
		{Key: "6", Title: "Android NDK Manager", Description: "Install or inspect C/C++ native Android NDK toolchains"},
		{Key: "7", Title: "Clean Caches & Reclaim Space", Description: "Purge Gradle build caches, daemon logs, and SDK intermediates"},
		{Key: "8", Title: "Shell Tab-Completion", Description: "Install or generate tab-completion for bash, zsh, and fish"},
		{Key: "9", Title: "Diagnostic Doctor", Description: "Comprehensive health & device environment diagnostics"},
		{Key: "10", Title: "Configuration Settings", Description: "Inspect or customize ~/.config/tadev/config.yml"},
		{Key: "11", Title: "Upgrade tadev CLI Binary", Description: "In-place self-updater from latest GitHub release"},
		{Key: "12", Title: "Uninstall Environment / Tool", Description: "Granular uninstallation chooser or complete removal"},
		{Key: "13", Title: "Exit Dashboard", Description: "Close and return to shell"},
	}

	idx, confirmed, err := ui.RunSelectMenu("Termux Android Development", cfg.DeviceClass, menuItems)
	if err != nil || !confirmed || idx == len(menuItems)-1 {
		return
	}

	switch idx {
	case 0:
		_ = cmd.RunSetup(false, "")
	case 1:
		_ = cmd.RunJavaCommand([]string{"switch"})
	case 2:
		_ = updater.ApplyUpdates(cfg)
	case 3:
		_ = updater.CheckUpdates(cfg, Version, false)
	case 4:
		_ = cmd.RunSdkManager([]string{"list"})
	case 5:
		_ = cmd.RunNdkCommand([]string{"status"})
	case 6:
		_ = cmd.RunCleanCommand(nil)
	case 7:
		_ = cmd.RunCompletionCommand([]string{"install"})
	case 8:
		_ = cmd.RunDoctor()
	case 9:
		_ = cmd.RunConfig([]string{"show"})
	case 10:
		_ = cmd.RunSelfUpdate(Version, false)
	case 11:
		_ = cmd.RunUninstall(cmd.UninstallFlags{})
	}
}

func main() {
	cfg, err := env.GetConfig()
	if err != nil {
		ui.Fail("Failed to read environment: " + err.Error())
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		if !cfg.IsTermux {
			ui.Fail("tadev is engineered exclusively for Termux on Android.")
			ui.Info("Cross-platform Termux architectures supported: ARM64, ARM (32-bit), x86_64, and x86.")
			ui.Info("Please run tadev within the Termux application.")
			os.Exit(1)
		}
		if ui.IsTerminal() {
			runDashboard(cfg)
			return
		}
		showHelp()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	if command == "version" || command == "--version" || command == "-v" {
		fmt.Printf("tadev %s\nAuthor: %s\nRepo:   %s\nDevice: %s\n", Version, Author, Repo, cfg.DeviceClass)
		return
	}

	if command == "help" || command == "--help" || command == "-h" {
		showHelp()
		return
	}

	hasFlag := func(flag string) bool {
		for _, a := range args {
			if a == flag {
				return true
			}
		}
		return false
	}

	getFlagValue := func(flag string) string {
		for i, a := range args {
			if (a == flag || a == "-"+flag) && i+1 < len(args) {
				return args[i+1]
			}
			if strings.HasPrefix(a, flag+"=") {
				return strings.TrimPrefix(a, flag+"=")
			}
		}
		return ""
	}

	if hasFlag("-h") || hasFlag("--help") {
		showHelp()
		return
	}

	if !cfg.IsTermux {
		ui.Fail("tadev is engineered exclusively for Termux on Android.")
		ui.Info("Cross-platform Termux architectures supported: ARM64, ARM (32-bit), x86_64, and x86.")
		ui.Info("Please run tadev within the Termux application.")
		os.Exit(1)
	}

	switch command {
	case "menu", "tui", "dashboard":
		runDashboard(cfg)

	case "install", "setup":
		force := hasFlag("--force")
		javaProfile := getFlagValue("--java")
		if javaProfile == "" {
			javaProfile = getFlagValue("--jdk")
		}
		if err := cmd.RunSetup(force, javaProfile); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "java", "jdk":
		if err := cmd.RunJavaCommand(args); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "uninstall", "remove", "desetup":
		flags := cmd.UninstallFlags{
			AutoConfirm:   hasFlag("-y") || hasFlag("--yes"),
			All:           hasFlag("--all"),
			Self:          hasFlag("--self"),
			SdkOnly:       hasFlag("--sdk-only"),
			PurgePackages: hasFlag("--purge-packages") || hasFlag("--packages"),
		}
		if err := cmd.RunUninstall(flags); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "update", "upgrade":
		if hasFlag("--cli") || hasFlag("--self") {
			if err := cmd.RunSelfUpdate(Version, false); err != nil {
				ui.Fail(err.Error())
				os.Exit(1)
			}
			return
		}
		if err := updater.ApplyUpdates(cfg); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "check", "check-updates":
		quiet := hasFlag("--quiet")
		if err := updater.CheckUpdates(cfg, Version, quiet); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "sdk":
		if err := cmd.RunSdkManager(args); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "ndk":
		if err := cmd.RunNdkCommand(args); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "clean", "clear", "purge-cache":
		if err := cmd.RunCleanCommand(args); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "completion", "autocomplete":
		if err := cmd.RunCompletionCommand(args); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "self-update", "self-upgrade":
		force := hasFlag("--force")
		if err := cmd.RunSelfUpdate(Version, force); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "doctor", "status", "health":
		if err := cmd.RunDoctor(); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	case "config", "cfg":
		if err := cmd.RunConfig(args); err != nil {
			ui.Fail(err.Error())
			os.Exit(1)
		}

	default:
		ui.Fail(fmt.Sprintf("Unknown command '%s'. Run 'tadev help' for usage.", command))
		os.Exit(1)
	}
}
