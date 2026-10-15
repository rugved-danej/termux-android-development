package env

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellConfigManagement(t *testing.T) {
	tempHome := t.TempDir()

	cfg := &Config{
		HomeDir:     tempHome,
		SdkDir:      filepath.Join(tempHome, ".android", "sdk"),
		JvmDir:      filepath.Join(tempHome, "jvm"),
		PrefixBin:   filepath.Join(tempHome, "bin"),
		BlockStart:  "# >>> termux-android-development >>>",
		BlockEnd:    "# <<< termux-android-development <<<",
		JavaProfile: "25+21",
	}

	// 1. Initially empty home dir - should at least identify bashrc and zshrc
	files := GetShellConfigFiles(cfg)
	if len(files) == 0 {
		t.Fatalf("expected at least default shell files, got none")
	}

	// 2. Configure shells
	primaryJava := filepath.Join(cfg.JvmDir, "java-25-openjdk")
	err := ConfigureShellWithJava(cfg, primaryJava)
	if err != nil {
		t.Fatalf("ConfigureShellWithJava failed: %v", err)
	}

	// Verify .bashrc was created and has the block
	bashrc := filepath.Join(tempHome, ".bashrc")
	bashData, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("expected .bashrc to exist: %v", err)
	}
	bashStr := string(bashData)
	if !strings.Contains(bashStr, cfg.BlockStart) || !strings.Contains(bashStr, "export JAVA_HOME=") {
		t.Fatalf("expected .bashrc to contain JAVA_HOME block, got:\n%s", bashStr)
	}
	if !strings.Contains(bashStr, "switch_java()") {
		t.Fatalf("expected .bashrc to contain switch_java() function, got:\n%s", bashStr)
	}
	if !strings.Contains(bashStr, "alias use-java=\"switch_java\"") {
		t.Fatalf("expected .bashrc to contain alias use-java, got:\n%s", bashStr)
	}

	// 3. Test fish configuration
	fishDir := filepath.Join(tempHome, ".config", "fish")
	_ = os.MkdirAll(fishDir, 0755)
	fishFile := filepath.Join(fishDir, "config.fish")
	_ = os.WriteFile(fishFile, []byte("# existing user fish config\n"), 0644)

	err = ConfigureShellWithJava(cfg, primaryJava)
	if err != nil {
		t.Fatalf("ConfigureShellWithJava with fish failed: %v", err)
	}
	fishData, err := os.ReadFile(fishFile)
	if err != nil {
		t.Fatalf("expected config.fish to exist: %v", err)
	}
	fishStr := string(fishData)
	if !strings.Contains(fishStr, "set -gx JAVA_HOME") {
		t.Fatalf("expected config.fish to contain fish-syntax 'set -gx JAVA_HOME', got:\n%s", fishStr)
	}
	if !strings.Contains(fishStr, "# existing user fish config") {
		t.Fatalf("expected config.fish to preserve existing user content")
	}

	// 4. Test CleanShell
	err = CleanShell(cfg)
	if err != nil {
		t.Fatalf("CleanShell failed: %v", err)
	}

	// Verify blocks were cleanly removed
	cleanedBash, _ := os.ReadFile(bashrc)
	if strings.Contains(string(cleanedBash), cfg.BlockStart) {
		t.Fatalf("CleanShell failed to remove block from .bashrc")
	}

	cleanedFish, _ := os.ReadFile(fishFile)
	if strings.Contains(string(cleanedFish), cfg.BlockStart) {
		t.Fatalf("CleanShell failed to remove block from config.fish")
	}
	if !strings.Contains(string(cleanedFish), "# existing user fish config") {
		t.Fatalf("CleanShell accidentally removed user content from config.fish")
	}
}

func TestShellReloadHint(t *testing.T) {
	cfg := &Config{HomeDir: "/test/home"}

	origShell := os.Getenv("SHELL")
	defer func() { _ = os.Setenv("SHELL", origShell) }()

	_ = os.Setenv("SHELL", "/bin/zsh")
	hint := GetShellReloadHint(cfg)
	if hint != "source ~/.zshrc" {
		t.Errorf("expected 'source ~/.zshrc' for zsh, got '%s'", hint)
	}

	_ = os.Setenv("SHELL", "/bin/bash")
	hint = GetShellReloadHint(cfg)
	if hint != "source ~/.bashrc" {
		t.Errorf("expected 'source ~/.bashrc' for bash, got '%s'", hint)
	}

	_ = os.Setenv("SHELL", "/usr/bin/fish")
	hint = GetShellReloadHint(cfg)
	if hint != "source ~/.config/fish/config.fish" {
		t.Errorf("expected fish hint, got '%s'", hint)
	}
}

func TestShellSyntaxValidity(t *testing.T) {
	tempHome := t.TempDir()

	cfg := &Config{
		HomeDir:     tempHome,
		SdkDir:      filepath.Join(tempHome, ".android", "sdk"),
		JvmDir:      filepath.Join(tempHome, "jvm"),
		PrefixBin:   filepath.Join(tempHome, "bin"),
		BlockStart:  "### BEGIN TADEV ###",
		BlockEnd:    "### END TADEV ###",
		JavaProfile: "21",
	}

	bashrc := filepath.Join(tempHome, ".bashrc")
	zshrc := filepath.Join(tempHome, ".zshrc")
	profile := filepath.Join(tempHome, ".profile")
	_ = os.WriteFile(bashrc, []byte(""), 0644)
	_ = os.WriteFile(zshrc, []byte(""), 0644)
	_ = os.WriteFile(profile, []byte(""), 0644)

	err := ConfigureShellWithJava(cfg, filepath.Join(cfg.JvmDir, "java-21-openjdk"))
	if err != nil {
		t.Fatalf("ConfigureShellWithJava failed: %v", err)
	}

	// 1. Validate bash syntax
	if _, err := exec.LookPath("bash"); err == nil {
		cmd := exec.Command("bash", "-n", bashrc)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("bash syntax validation failed on generated .bashrc:\n%s", string(out))
		}
	}

	// 2. Validate zsh syntax & ensure double-sourcing produces zero errors
	if _, err := exec.LookPath("zsh"); err == nil {
		cmd := exec.Command("zsh", "-n", zshrc)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("zsh syntax validation failed on generated .zshrc:\n%s", string(out))
		}

		// Double-source test in zsh
		cmdSource := exec.Command("zsh", "-c", "source "+zshrc+" && source "+zshrc)
		if out, err := cmdSource.CombinedOutput(); err != nil || len(out) > 0 {
			t.Errorf("zsh double-sourcing generated output or error: err=%v, out=%s", err, string(out))
		}
	}

	// 3. Validate dash syntax (standard /bin/sh POSIX on Debian/Ubuntu)
	if _, err := exec.LookPath("dash"); err == nil {
		cmd := exec.Command("dash", "-n", profile)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("dash syntax validation failed on generated .profile:\n%s", string(out))
		}
	}
}
