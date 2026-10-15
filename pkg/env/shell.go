package env

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// GetShellConfigFiles returns all detected and candidate shell RC files to configure.
// Discovers active shell, installed shell binaries (bash, zsh, fish, ksh, csh), and existing configs.
func GetShellConfigFiles(cfg *Config) []string {
	var files []string
	seen := make(map[string]bool)

	add := func(p string) {
		if p != "" && !seen[p] {
			seen[p] = true
			files = append(files, p)
		}
	}

	bashrc := filepath.Join(cfg.HomeDir, ".bashrc")
	zshrc := filepath.Join(cfg.HomeDir, ".zshrc")
	bashProfile := filepath.Join(cfg.HomeDir, ".bash_profile")
	zprofile := filepath.Join(cfg.HomeDir, ".zprofile")
	profile := filepath.Join(cfg.HomeDir, ".profile")
	fishConfig := filepath.Join(cfg.HomeDir, ".config", "fish", "config.fish")
	mkshrc := filepath.Join(cfg.HomeDir, ".mkshrc")
	kshrc := filepath.Join(cfg.HomeDir, ".kshrc")
	cshrc := filepath.Join(cfg.HomeDir, ".cshrc")
	tcshrc := filepath.Join(cfg.HomeDir, ".tcshrc")

	// 1. Any candidate config file that already exists on disk
	candidates := []string{bashrc, zshrc, bashProfile, zprofile, profile, fishConfig, mkshrc, kshrc, cshrc, tcshrc}
	for _, cand := range candidates {
		if _, err := os.Stat(cand); err == nil {
			add(cand)
		}
	}

	// Helper to check if a shell binary exists on system
	hasBinary := func(names ...string) bool {
		for _, name := range names {
			if _, err := exec.LookPath(name); err == nil {
				return true
			}
			if _, err := os.Stat(filepath.Join(cfg.PrefixBin, name)); err == nil {
				return true
			}
			if _, err := os.Stat(filepath.Join("/bin", name)); err == nil {
				return true
			}
			if _, err := os.Stat(filepath.Join("/usr/bin", name)); err == nil {
				return true
			}
		}
		return false
	}

	activeShell := os.Getenv("SHELL")

	// 2. Proactively configure primary RC files for installed or active shells
	// Bash: Default in Termux and Linux
	if hasBinary("bash") || strings.Contains(activeShell, "bash") || len(files) == 0 {
		add(bashrc)
	}

	// Zsh: Highly popular in Termux and Linux
	if hasBinary("zsh") || strings.Contains(activeShell, "zsh") {
		add(zshrc)
	}

	// Fish
	if hasBinary("fish") || strings.Contains(activeShell, "fish") {
		add(fishConfig)
	}

	// Mksh / Ksh (Android native or user ksh)
	if hasBinary("mksh") || strings.Contains(activeShell, "mksh") {
		add(mkshrc)
	}
	if hasBinary("ksh") || strings.Contains(activeShell, "ksh") {
		add(kshrc)
	}

	// Csh / Tcsh
	if hasBinary("csh") || hasBinary("tcsh") || strings.Contains(activeShell, "csh") || strings.Contains(activeShell, "tcsh") {
		add(cshrc)
	}

	// Fallback to ensure at least bash and zsh are present
	if len(files) == 0 {
		add(bashrc)
		add(zshrc)
	}

	return files
}

// GetShellReloadHint provides a shell-aware command suggestion (e.g. 'source ~/.bashrc' or 'source ~/.zshrc').
func GetShellReloadHint(cfg *Config) string {
	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return "source ~/.zshrc"
	case strings.Contains(shell, "fish"):
		return "source ~/.config/fish/config.fish"
	case strings.Contains(shell, "bash"):
		return "source ~/.bashrc"
	case strings.Contains(shell, "csh") || strings.Contains(shell, "tcsh"):
		return "source ~/.cshrc"
	case strings.Contains(shell, "mksh"):
		return ". ~/.mkshrc"
	default:
		if _, err := os.Stat(filepath.Join(cfg.HomeDir, ".bashrc")); err == nil {
			return "source ~/.bashrc"
		}
		if _, err := os.Stat(filepath.Join(cfg.HomeDir, ".zshrc")); err == nil {
			return "source ~/.zshrc"
		}
		return "source ~/.bashrc"
	}
}

// ConfigureShell updates all detected shell RC files with the active Java and Android environment.
func ConfigureShell(cfg *Config) error {
	return ConfigureShellWithJava(cfg, cfg.GetPrimaryJavaHome())
}

// ConfigureShellWithJava writes the shell environment blocks for bash, zsh, fish, ksh, and csh.
func ConfigureShellWithJava(cfg *Config, primaryJava string) error {
	javaOpts := cfg.GetJavaOpts()
	ndkPath := GetActiveNdkPath(cfg)

	ndkPosix := ""
	if ndkPath != "" {
		ndkPosix = fmt.Sprintf("\nexport ANDROID_NDK_HOME=\"%s\"\nexport NDK_HOME=\"$ANDROID_NDK_HOME\"\nexport PATH=\"$ANDROID_NDK_HOME:$PATH\"", ndkPath)
	}

	ndkFish := ""
	if ndkPath != "" {
		ndkFish = fmt.Sprintf("\nset -gx ANDROID_NDK_HOME \"%s\"\nset -gx NDK_HOME \"$ANDROID_NDK_HOME\"\nset -gx PATH $ANDROID_NDK_HOME $PATH", ndkPath)
	}

	ndkCsh := ""
	if ndkPath != "" {
		ndkCsh = fmt.Sprintf("\nsetenv ANDROID_NDK_HOME \"%s\"\nsetenv NDK_HOME \"%s\"\nset path = ( %s $path )", ndkPath, ndkPath, ndkPath)
	}

	posixBlock := fmt.Sprintf(`%s
# ==========================================
# Java & Android SDK Development Environment
# ==========================================
export JAVA_HOME="%s"
export ANDROID_HOME="%s"
export ANDROID_SDK_ROOT="%s"
export PATH="$JAVA_HOME/bin:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$PATH"
export JAVA_OPTS="%s"%s

# Android development shortcuts
alias gw="bash ./gradlew"
alias gradlew="bash ./gradlew"
alias sdk="sdkmanager"

# Helper to switch Java versions seamlessly ('use-java 21' or 'switch_java 21')
switch_java() {
  ver="$1"
  if [ -z "$ver" ]; then
    if command -v tadev >/dev/null 2>&1; then
      tadev java list
    else
      echo "Usage: use-java <version>"
      echo "Installed versions in %s:"
      command ls -1 %s 2>/dev/null | command grep -E "java-[0-9]+"
    fi
    return 1
  fi
  if command -v tadev >/dev/null 2>&1; then
    tadev java switch "$ver" || return $?
    host_ver="${ver%%%%+*}"
    if [ -d "%s/java-${host_ver}-openjdk" ]; then
      export JAVA_HOME="%s/java-${host_ver}-openjdk"
      export PATH="$JAVA_HOME/bin:$PATH"
    fi
    return 0
  fi
  if [ -d "%s/java-${ver}-openjdk" ]; then
    export JAVA_HOME="%s/java-${ver}-openjdk"
    export PATH="$JAVA_HOME/bin:$PATH"
    echo "Switched to OpenJDK ${ver}: $JAVA_HOME"
    java -version
  else
    echo "Installed versions in %s:"
    command ls -1 %s 2>/dev/null | command grep -E "java-[0-9]+"
  fi
}

use_java() {
  switch_java "$@"
}
alias use-java="switch_java"

# Auto-check for updates (runs non-intrusively in background once every 7 days)
if [ -x "%s/tadev" ]; then
  (tadev check --quiet &) 2>/dev/null
fi
%s
`, cfg.BlockStart, primaryJava, cfg.SdkDir, cfg.SdkDir, javaOpts, ndkPosix, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.PrefixBin, cfg.BlockEnd)

	fishBlock := fmt.Sprintf(`%s
# ==========================================
# Java & Android SDK Development Environment
# ==========================================
set -gx JAVA_HOME "%s"
set -gx ANDROID_HOME "%s"
set -gx ANDROID_SDK_ROOT "%s"
set -gx PATH $JAVA_HOME/bin $ANDROID_HOME/cmdline-tools/latest/bin $ANDROID_HOME/platform-tools $PATH
set -gx JAVA_OPTS "%s"%s

alias gw="bash ./gradlew"
alias gradlew="bash ./gradlew"
alias sdk="sdkmanager"

function switch_java
    set -l ver $argv[1]
    if test -z "$ver"
        if command -sq tadev
            tadev java list
        else
            echo "Usage: use-java <version>"
            echo "Installed versions in %s:"
            command ls -1 %s 2>/dev/null | command grep -E "java-[0-9]+"
        end
        return 1
    end
    if command -sq tadev
        tadev java switch $ver
        or return $status
        set -l host_ver (string replace -r '\+.*' '' $ver)
        if test -d "%s/java-$host_ver-openjdk"
            set -gx JAVA_HOME "%s/java-$host_ver-openjdk"
            set -gx PATH "$JAVA_HOME/bin" $PATH
        end
        return 0
    end
    if test -d "%s/java-$ver-openjdk"
        set -gx JAVA_HOME "%s/java-$ver-openjdk"
        set -gx PATH "$JAVA_HOME/bin" $PATH
        echo "Switched to OpenJDK $ver: $JAVA_HOME"
        java -version
    else
        echo "Installed versions in %s:"
        command ls -1 %s 2>/dev/null | command grep -E "java-[0-9]+"
    end
end

function use-java
    switch_java $argv
end

if test -x "%s/tadev"
    tadev check --quiet &>/dev/null &
end
%s
`, cfg.BlockStart, primaryJava, cfg.SdkDir, cfg.SdkDir, javaOpts, ndkFish, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.JvmDir, cfg.PrefixBin, cfg.BlockEnd)

	cshBlock := fmt.Sprintf(`%s
# ==========================================
# Java & Android SDK Development Environment
# ==========================================
setenv JAVA_HOME "%s"
setenv ANDROID_HOME "%s"
setenv ANDROID_SDK_ROOT "%s"
setenv JAVA_OPTS "%s"
set path = ( %s/bin %s/cmdline-tools/latest/bin %s/platform-tools $path )%s

alias gw "bash ./gradlew"
alias gradlew "bash ./gradlew"
alias sdk "sdkmanager"
%s
`, cfg.BlockStart, primaryJava, cfg.SdkDir, cfg.SdkDir, javaOpts, primaryJava, cfg.SdkDir, cfg.SdkDir, ndkCsh, cfg.BlockEnd)

	files := GetShellConfigFiles(cfg)
	for _, f := range files {
		content := ""
		if data, err := os.ReadFile(f); err == nil {
			content = string(data)
		}
		content = RemoveManagedBlock(content, cfg.BlockStart, cfg.BlockEnd)
		newContent := strings.TrimSpace(content)
		if newContent != "" {
			newContent += "\n\n"
		}

		switch {
		case strings.HasSuffix(f, ".fish"):
			_ = os.MkdirAll(filepath.Dir(f), 0755)
			newContent += fishBlock
		case strings.HasSuffix(f, "cshrc"):
			_ = os.MkdirAll(filepath.Dir(f), 0755)
			newContent += cshBlock
		default:
			// bash, zsh, ksh, mksh, profile
			_ = os.MkdirAll(filepath.Dir(f), 0755)
			newContent += posixBlock
		}

		if err := os.WriteFile(f, []byte(newContent), 0644); err != nil {
			return err
		}
	}
	return nil
}

// CleanShell removes all managed environment blocks from all candidate shell RC files.
func CleanShell(cfg *Config) error {
	files := GetShellConfigFiles(cfg)
	for _, f := range files {
		if data, err := os.ReadFile(f); err == nil {
			cleaned := RemoveManagedBlock(string(data), cfg.BlockStart, cfg.BlockEnd)
			cleaned = strings.TrimSpace(cleaned)
			if cleaned != "" {
				cleaned += "\n"
			}
			_ = os.WriteFile(f, []byte(cleaned), 0644)
		}
	}
	return nil
}

// ClearAndRestartShell clears the console and automatically restarts the user's shell session.
func ClearAndRestartShell(cfg *Config) error {
	// Clear the terminal screen using ANSI escape sequences
	fmt.Print("\033[H\033[2J\033[3J")

	shell := os.Getenv("SHELL")
	if shell == "" {
		if _, err := os.Stat(filepath.Join(cfg.PrefixBin, "bash")); err == nil {
			shell = filepath.Join(cfg.PrefixBin, "bash")
		} else if _, err := os.Stat("/bin/sh"); err == nil {
			shell = "/bin/sh"
		} else {
			shell = "sh"
		}
	}

	shellPath, err := exec.LookPath(shell)
	if err != nil {
		shellPath = shell
	}

	return syscall.Exec(shellPath, []string{filepath.Base(shellPath), "-l"}, os.Environ())
}
