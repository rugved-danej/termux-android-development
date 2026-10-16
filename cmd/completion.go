package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

// RunCompletionCommand outputs or installs native tab-completion scripts for bash, zsh, and fish.
func RunCompletionCommand(args []string) error {
	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	target := ""
	if len(args) > 0 {
		target = strings.ToLower(args[0])
	}

	switch target {
	case "bash":
		fmt.Print(BashCompletionScript)
		return nil

	case "zsh":
		fmt.Print(ZshCompletionScript)
		return nil

	case "fish":
		fmt.Print(FishCompletionScript)
		return nil

	case "install":
		return installCompletion(cfg)

	default:
		fmt.Printf("%s\n\n", ui.BrightPurple("◆ Shell Tab-Completion for tadev"))
		fmt.Println("Generate or install tab-completion for your active shell.")
		fmt.Println()
		fmt.Printf("%s\n", ui.BrightPurple("◆ USAGE"))
		fmt.Printf("  %s %s\n", ui.Purple("tadev completion bash"), ui.GrayText("# Print Bash completion script"))
		fmt.Printf("  %s %s\n", ui.Purple("tadev completion zsh"), ui.GrayText("# Print Zsh completion script"))
		fmt.Printf("  %s %s\n", ui.Purple("tadev completion fish"), ui.GrayText("# Print Fish completion script"))
		fmt.Printf("  %s %s\n\n", ui.Purple("tadev completion install"), ui.GrayText("# Automatically install completion into your shell RC"))
		fmt.Printf("%s\n", ui.BrightPurple("◆ QUICK SETUP"))
		fmt.Printf("  %s  %s\n", ui.Purple("Bash:"), ui.YellowText("echo 'source <(tadev completion bash)' >> ~/.bashrc"))
		fmt.Printf("  %s   %s\n", ui.Purple("Zsh:"), ui.YellowText("echo 'source <(tadev completion zsh)' >> ~/.zshrc"))
		fmt.Printf("  %s  %s\n\n", ui.Purple("Fish:"), ui.YellowText("tadev completion fish > ~/.config/fish/completions/tadev.fish"))
		return nil
	}
}

func installCompletion(cfg *env.Config) error {
	shell := os.Getenv("SHELL")
	sp := ui.NewSpinner("Installing shell tab-completion...")
	sp.Start()

	switch {
	case strings.Contains(shell, "fish"):
		fishCompDir := filepath.Join(cfg.HomeDir, ".config", "fish", "completions")
		_ = os.MkdirAll(fishCompDir, 0755)
		fishCompFile := filepath.Join(fishCompDir, "tadev.fish")
		if err := os.WriteFile(fishCompFile, []byte(FishCompletionScript), 0644); err != nil {
			sp.StopFail("Failed to write fish completion: " + err.Error())
			return err
		}
		sp.StopSuccess(fmt.Sprintf("Fish tab-completion installed at %s", fishCompFile))

	case strings.Contains(shell, "zsh"):
		zshrc := filepath.Join(cfg.HomeDir, ".zshrc")
		block := "\n# tadev zsh completion\nif command -v tadev >/dev/null 2>&1; then\n  source <(tadev completion zsh)\nfi\n"
		if data, err := os.ReadFile(zshrc); err == nil && !strings.Contains(string(data), "tadev completion zsh") {
			f, err := os.OpenFile(zshrc, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				_, _ = f.WriteString(block)
				f.Close()
			}
		}
		sp.StopSuccess(fmt.Sprintf("Zsh tab-completion hook added to %s", zshrc))

	default:
		// Default to bash
		bashrc := filepath.Join(cfg.HomeDir, ".bashrc")
		block := "\n# tadev bash completion\nif command -v tadev >/dev/null 2>&1; then\n  source <(tadev completion bash)\nfi\n"
		if data, err := os.ReadFile(bashrc); err == nil && !strings.Contains(string(data), "tadev completion bash") {
			f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				_, _ = f.WriteString(block)
				f.Close()
			}
		}
		sp.StopSuccess(fmt.Sprintf("Bash tab-completion hook added to %s", bashrc))
	}

	fmt.Printf("\nRestart your shell or run '%s' to enable tab-completion.\n\n", ui.Purple(env.GetShellReloadHint(cfg)))
	return nil
}

const BashCompletionScript = `# bash completion for tadev & termux-android-dev
_tadev_completions() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    local commands="install setup java jdk uninstall remove desetup update upgrade check check-updates sdk ndk clean doctor status config cfg self-update menu tui dashboard completion version help"
    local java_profiles="25+21 21 26+21 26 17 25+17"

    if [ "$COMP_CWORD" -eq 1 ]; then
        COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
        return 0
    fi

    case "$prev" in
        java|jdk)
            COMPREPLY=( $(compgen -W "switch list test status $java_profiles" -- "$cur") )
            return 0
            ;;
        switch)
            COMPREPLY=( $(compgen -W "$java_profiles" -- "$cur") )
            return 0
            ;;
        sdk)
            COMPREPLY=( $(compgen -W "list install" -- "$cur") )
            return 0
            ;;
        ndk)
            COMPREPLY=( $(compgen -W "status list install remove" -- "$cur") )
            return 0
            ;;
        clean)
            COMPREPLY=( $(compgen -W "all gradle sdk tmp --all --gradle --sdk --tmp -y" -- "$cur") )
            return 0
            ;;
        config|cfg)
            COMPREPLY=( $(compgen -W "show init reset path" -- "$cur") )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish install" -- "$cur") )
            return 0
            ;;
        --java|--jdk)
            COMPREPLY=( $(compgen -W "$java_profiles" -- "$cur") )
            return 0
            ;;
    esac

    if [[ "$cur" == -* ]]; then
        local flags="--java --jdk --all --self --packages --sdk-only --force --quiet --cli -y --yes -h --help -v --version"
        COMPREPLY=( $(compgen -W "$flags" -- "$cur") )
        return 0
    fi
}
complete -F _tadev_completions tadev termux-android-dev
`

const ZshCompletionScript = `#compdef tadev termux-android-dev

_tadev() {
    local -a commands
    commands=(
        'install:Install complete Android development environment'
        'setup:Install complete Android development environment'
        'java:Manage OpenJDK profiles (25+21, 21, 26+21, 26, 17)'
        'jdk:Manage OpenJDK profiles'
        'uninstall:Granular uninstallation or complete removal'
        'remove:Granular uninstallation'
        'update:Download and apply all updates'
        'upgrade:Download and apply all updates'
        'check:Check for updates across repos and GitHub'
        'check-updates:Check for updates'
        'sdk:Inspect installed platforms or install extra SDK packages'
        'ndk:Manage Android NDK installations and toolchains'
        'clean:Purge temporary build caches and daemons'
        'doctor:Run diagnostic health checks on all components'
        'status:Run diagnostic health checks'
        'config:Inspect or initialize configuration'
        'self-update:Upgrade tadev binary directly from GitHub'
        'menu:Launch interactive TUI dashboard'
        'tui:Launch interactive TUI dashboard'
        'completion:Generate shell autocompletion script'
        'version:Show version information'
        'help:Show help guide'
    )

    local -a java_profiles
    java_profiles=(
        '25+21:OpenJDK 25 (Host) + OpenJDK 21 (Toolchain) [Default]'
        '21:OpenJDK 21 LTS Only (Space Saver, 100% AGP Compatible)'
        '26+21:OpenJDK 26 (TUR Host) + OpenJDK 21 (Toolchain)'
        '26:OpenJDK 26 Cutting-Edge Only (TUR Repo)'
        '17:OpenJDK 17 LTS Only (AGP 8.0 Baseline)'
        '25+17:OpenJDK 25 (Host) + OpenJDK 17 (Toolchain)'
    )

    if (( CURRENT == 2 )); then
        _describe -t commands 'tadev commands' commands
        return
    fi

    case $words[2] in
        java|jdk)
            if (( CURRENT == 3 )); then
                local -a java_cmds
                java_cmds=('switch:Switch active profile' 'list:List installed JDKs' 'test:Test runtime compatibility')
                _describe -t java_cmds 'java commands' java_cmds
                _describe -t java_profiles 'profiles' java_profiles
            elif [[ $words[3] == "switch" && CURRENT == 4 ]]; then
                _describe -t java_profiles 'profiles' java_profiles
            fi
            ;;
        sdk)
            if (( CURRENT == 3 )); then
                local -a sdk_cmds=('list:List installed platforms' 'install:Install extra SDK package')
                _describe -t sdk_cmds 'sdk commands' sdk_cmds
            fi
            ;;
        ndk)
            if (( CURRENT == 3 )); then
                local -a ndk_cmds=('status:Show active NDK' 'list:List installed NDKs' 'install:Install Android NDK' 'remove:Remove NDK')
                _describe -t ndk_cmds 'ndk commands' ndk_cmds
            fi
            ;;
        clean)
            if (( CURRENT == 3 )); then
                local -a clean_cmds=('all:Clean all caches' 'gradle:Clean Gradle caches' 'sdk:Clean SDK caches' 'tmp:Clean temporary files')
                _describe -t clean_cmds 'clean targets' clean_cmds
            fi
            ;;
        config|cfg)
            if (( CURRENT == 3 )); then
                local -a cfg_cmds=('show:View active YAML' 'init:Create config file' 'reset:Reset to defaults' 'path:Print config path')
                _describe -t cfg_cmds 'config actions' cfg_cmds
            fi
            ;;
        completion)
            if (( CURRENT == 3 )); then
                local -a shells=('bash:Bash completion' 'zsh:Zsh completion' 'fish:Fish completion' 'install:Auto-install to shell RC')
                _describe -t shells 'supported shells' shells
            fi
            ;;
    esac
}

_tadev "$@"
`

const FishCompletionScript = `# fish completion for tadev & termux-android-dev
complete -c tadev -f
complete -c termux-android-dev -f

# Main commands
complete -c tadev -n "__fish_use_subcommand" -a install -d "Install complete Android development environment"
complete -c tadev -n "__fish_use_subcommand" -a setup -d "Install complete Android development environment"
complete -c tadev -n "__fish_use_subcommand" -a java -d "Manage OpenJDK profiles (25+21, 21, 26+21, 26, 17)"
complete -c tadev -n "__fish_use_subcommand" -a jdk -d "Manage OpenJDK profiles"
complete -c tadev -n "__fish_use_subcommand" -a uninstall -d "Granular uninstallation or complete removal"
complete -c tadev -n "__fish_use_subcommand" -a remove -d "Granular uninstallation"
complete -c tadev -n "__fish_use_subcommand" -a update -d "Download and apply all updates"
complete -c tadev -n "__fish_use_subcommand" -a upgrade -d "Download and apply all updates"
complete -c tadev -n "__fish_use_subcommand" -a check -d "Check for updates across repos and GitHub"
complete -c tadev -n "__fish_use_subcommand" -a check-updates -d "Check for updates"
complete -c tadev -n "__fish_use_subcommand" -a sdk -d "Inspect installed platforms or install extra SDK packages"
complete -c tadev -n "__fish_use_subcommand" -a ndk -d "Manage Android NDK installations and toolchains"
complete -c tadev -n "__fish_use_subcommand" -a clean -d "Purge temporary build caches and daemons"
complete -c tadev -n "__fish_use_subcommand" -a doctor -d "Run diagnostic health checks on all components"
complete -c tadev -n "__fish_use_subcommand" -a status -d "Run diagnostic health checks"
complete -c tadev -n "__fish_use_subcommand" -a config -d "Inspect or initialize configuration"
complete -c tadev -n "__fish_use_subcommand" -a self-update -d "Upgrade tadev binary directly from GitHub"
complete -c tadev -n "__fish_use_subcommand" -a menu -d "Launch interactive TUI dashboard"
complete -c tadev -n "__fish_use_subcommand" -a tui -d "Launch interactive TUI dashboard"
complete -c tadev -n "__fish_use_subcommand" -a completion -d "Generate shell autocompletion script"
complete -c tadev -n "__fish_use_subcommand" -a version -d "Show version information"
complete -c tadev -n "__fish_use_subcommand" -a help -d "Display help guide"

# Subcommands
complete -c tadev -n "__fish_seen_subcommand_from java jdk" -a "switch list test 25+21 21 26+21 26 17 25+17"
complete -c tadev -n "__fish_seen_subcommand_from sdk" -a "list install"
complete -c tadev -n "__fish_seen_subcommand_from ndk" -a "status list install remove"
complete -c tadev -n "__fish_seen_subcommand_from clean" -a "all gradle sdk tmp"
complete -c tadev -n "__fish_seen_subcommand_from config cfg" -a "show init reset path"
complete -c tadev -n "__fish_seen_subcommand_from completion" -a "bash zsh fish install"

# Options
complete -c tadev -l java -a "25+21 21 26+21 26 17 25+17" -d "Java profile"
complete -c tadev -l force -d "Force re-installation"
complete -c tadev -l quiet -d "Quiet mode"
complete -c tadev -l all -d "Purge all components"
complete -c tadev -l self -d "Self-remove tadev tool"
complete -c tadev -s y -l yes -d "Skip confirmation"
`
