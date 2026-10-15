package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

const DefaultConfigYAML = `# ==============================================================================
#  tadev - Termux Android Development Engine Configuration
#  Author: Rugved Danej (@rugved-danej)
#  Repository: https://github.com/rugved-danej/termux-android-development
# ==============================================================================

version: "1.0.0"
author: "Rugved Danej"
repository: "https://github.com/rugved-danej/termux-android-development"

# Android SDK Environment Configuration
android:
  sdk_root: "~/.android/sdk"
  cmdline_tools:
    auto_detect_latest: true
    fallback_build: "15641748"
  platforms:
    - "android-36"
    - "android-35"
  build_tools:
    - "36.0.0"
  aapt2:
    native_override: true
    binary_path: "/data/data/com.termux/files/usr/bin/aapt2"

# OpenJDK & Java Toolchain Settings
java:
  profile: "25+21" # Options: "25+21", "21", "26+21", "26", "17", "25+17"
  default_version: "25"
  default_path: "/data/data/com.termux/files/usr/lib/jvm/java-25-openjdk"
  toolchain_paths:
    - "/data/data/com.termux/files/usr/lib/jvm/java-25-openjdk"
    - "/data/data/com.termux/files/usr/lib/jvm/java-21-openjdk"
  jvm_args: "-Xmx2048m -XX:MaxMetaspaceSize=512m -XX:+UseG1GC"

# Terminal UI & Animation Preferences
ui:
  animations_enabled: true
  spinner_style: "braille"
  spinner_interval_ms: 75
  show_elapsed_timer: true
  color_theme: "cyberpunk-cyan"
  show_banner: false

# Update & Health-Check Automation
updates:
  check_frequency_days: 7
  auto_check_in_shell: true
  quiet_mode: false
`

func RunConfig(args []string) error {
	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	subcommand := "show"
	if len(args) > 0 {
		subcommand = strings.ToLower(args[0])
	}

	switch subcommand {
	case "path":
		fmt.Println(cfg.ConfigFile)
		return nil

	case "init":
		if err := os.MkdirAll(cfg.ConfigDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if _, err := os.Stat(cfg.ConfigFile); err == nil {
			ui.Warn(fmt.Sprintf("Config file already exists at %s", cfg.ConfigFile))
			ui.Info("Use 'tadev config reset' to overwrite with fresh defaults.")
			return nil
		}
		if err := os.WriteFile(cfg.ConfigFile, []byte(DefaultConfigYAML), 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		ui.Success(fmt.Sprintf("Initialized tadev configuration at %s", cfg.ConfigFile))
		return nil

	case "reset":
		if err := os.MkdirAll(cfg.ConfigDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if err := os.WriteFile(cfg.ConfigFile, []byte(DefaultConfigYAML), 0644); err != nil {
			return fmt.Errorf("failed to reset config file: %w", err)
		}
		ui.Success(fmt.Sprintf("Reset configuration to defaults at %s", cfg.ConfigFile))
		return nil

	case "show", "cat", "view":
		ui.PrintBanner()
		content := ""
		if data, err := os.ReadFile(cfg.ConfigFile); err == nil {
			content = string(data)
			fmt.Printf("%sConfiguration File: %s%s%s\n\n", ui.Bold+ui.Yellow, ui.Cyan, cfg.ConfigFile, ui.Reset)
		} else {
			content = DefaultConfigYAML
			fmt.Printf("%sActive Configuration (Built-in Defaults - File not yet created)%s\n", ui.Bold+ui.Yellow, ui.Reset)
			fmt.Printf("%sRun 'tadev config init' to persist to %s%s\n\n", ui.Gray, cfg.ConfigFile, ui.Reset)
		}

		// Colorize YAML syntax for terminal viewing
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				fmt.Println(ui.Gray + line + ui.Reset)
			} else if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				key := parts[0]
				val := parts[1]
				fmt.Printf("%s%s:%s%s%s\n", ui.Cyan+ui.Bold, key, ui.Reset, ui.Green, val+ui.Reset)
			} else {
				fmt.Println(line)
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown config action '%s'. Valid actions: show, init, reset, path", subcommand)
	}
}

// EnsureConfigFile ensures that ~/.config/tadev/config.yml exists.
func EnsureConfigFile(cfg *env.Config) {
	if _, err := os.Stat(cfg.ConfigFile); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(cfg.ConfigFile), 0755)
		_ = os.WriteFile(cfg.ConfigFile, []byte(DefaultConfigYAML), 0644)
	}
}
