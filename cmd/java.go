package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

// RunJavaCommand handles java profile management and runtime diagnostics.
func RunJavaCommand(args []string) error {
	ui.PrintBanner()

	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	subcommand := "menu"
	if len(args) > 0 {
		subcommand = strings.ToLower(args[0])
	}

	switch subcommand {
	case "list", "status", "show":
		return runJavaList(cfg)
	case "test", "check":
		return runJavaTest(cfg)
	case "switch", "select", "set":
		target := ""
		if len(args) > 1 {
			target = args[1]
		}
		return runJavaSwitch(cfg, target)
	default:
		// If argument is a profile name directly (e.g. 'tadev java 21' or 'tadev java 26+21')
		if _, ok := env.GetJavaProfile(subcommand); ok {
			return runJavaSwitch(cfg, subcommand)
		}
		return runJavaSwitch(cfg, "")
	}
}

func runJavaList(cfg *env.Config) error {
	activeProfile := cfg.GetConfiguredJavaProfile()

	fmt.Printf("%s\n\n", ui.BrightPurple("◆ Configured Java Profile"))
	ui.Infof("Active Profile ID: %s", ui.Purple(activeProfile.ID))
	ui.Info(fmt.Sprintf("Title:             %s", activeProfile.Title))
	ui.Info(fmt.Sprintf("Host Directory:    %s", filepath.Join(cfg.JvmDir, activeProfile.HostJavaDir)))
	ui.Info(fmt.Sprintf("Toolchain Paths:   %s", strings.Join(activeProfile.ToolchainDirs, ", ")))

	fmt.Printf("\n%s\n\n", ui.BrightPurple("◆ Installed OpenJDK Runtimes"))
	installed := env.DetectInstalledJavaVersions(cfg)
	if len(installed) == 0 {
		ui.Warn(fmt.Sprintf("No OpenJDK installations detected in %s", cfg.JvmDir))
	} else {
		for _, info := range installed {
			activeTag := ""
			if info.IsActive {
				activeTag = fmt.Sprintf(" %s", ui.GreenText("[ACTIVE IN PATH]"))
			}
			ui.Success(fmt.Sprintf("OpenJDK %d: %s (%s)%s", info.Major, info.VersionLine, info.Path, activeTag))
		}
	}

	fmt.Printf("\n%s\n\n", ui.BrightPurple("◆ Available Profiles"))
	for _, p := range env.AvailableJavaProfiles {
		marker := "  "
		if p.ID == activeProfile.ID {
			marker = fmt.Sprintf("%s ", ui.Purple(ui.SymbolArrow))
		}
		ui.Info(fmt.Sprintf("%s%s", marker, p.Title))
	}
	fmt.Println()
	return nil
}

func runJavaSwitch(cfg *env.Config, targetID string) error {
	var targetProfile env.JavaProfile
	if targetID != "" {
		p, ok := env.GetJavaProfile(targetID)
		if !ok {
			return fmt.Errorf("unknown Java profile '%s'. Run 'tadev java list' to view supported profiles", targetID)
		}
		targetProfile = p
	} else if ui.IsTerminal() {
		var menuItems []ui.MenuItem
		for _, p := range env.AvailableJavaProfiles {
			menuItems = append(menuItems, ui.MenuItem{
				Key:   p.ID,
				Title: p.Title,
			})
		}

		idx, confirmed, err := ui.RunJavaProfileMenu("SWITCH JAVA PROFILE", cfg.DeviceClass, menuItems, env.JavaProfilesDocsURL)
		if err != nil || !confirmed || idx < 0 || idx >= len(env.AvailableJavaProfiles) {
			ui.Info("Profile switch cancelled.")
			return nil
		}
		targetProfile = env.AvailableJavaProfiles[idx]
	} else {
		targetProfile = cfg.GetConfiguredJavaProfile()
	}

	sp := ui.NewSpinner(fmt.Sprintf("Switching Java profile to %s...", targetProfile.ID))
	sp.Start()

	if err := env.ApplyJavaProfile(cfg, targetProfile); err != nil {
		sp.StopFail(err.Error())
		return err
	}

	sp.StopSuccess(fmt.Sprintf("Java profile switched to %s successfully.", targetProfile.ID))

	fmt.Println()
	summary := []string{
		fmt.Sprintf("Active Profile:   %s", ui.GreenText(targetProfile.ID)),
		fmt.Sprintf("Title:            %s", targetProfile.Title),
		fmt.Sprintf("Host OpenJDK:     %s", filepath.Join(cfg.JvmDir, targetProfile.HostJavaDir)),
		fmt.Sprintf("Toolchain Paths:  %s", strings.Join(targetProfile.ToolchainDirs, ", ")),
		fmt.Sprintf("Requires TUR:     %t", targetProfile.RequiresTur),
	}
	ui.PrintBox("JAVA PROFILE UPDATED", summary)

	if ui.IsTerminal() {
		fmt.Printf("\n%s\n", ui.Purple("Clearing console and restarting shell in 2 seconds..."))
		time.Sleep(2 * time.Second)
		return env.ClearAndRestartShell(cfg)
	}

	fmt.Printf("\nRestart your shell or run '%s' to apply environment variables.\n\n", ui.Purple(env.GetShellReloadHint(cfg)))
	return nil
}

func runJavaTest(cfg *env.Config) error {
	fmt.Printf("%s\n\n", ui.BrightPurple("◆ Testing All Installed OpenJDK Versions & Build Tool Compatibility"))

	installed := env.DetectInstalledJavaVersions(cfg)
	if len(installed) == 0 {
		ui.Fail(fmt.Sprintf("No OpenJDK installations detected in %s", cfg.JvmDir))
		return nil
	}

	for _, info := range installed {
		fmt.Printf("%s\n", ui.Purple(fmt.Sprintf("--- OpenJDK %d (%s) ---", info.Major, info.DirName)))

		// 1. Runtime Version
		binJava := filepath.Join(info.Path, "bin", "java")
		if out, err := exec.Command(binJava, "-version").CombinedOutput(); err == nil {
			firstLine := strings.Split(string(out), "\n")[0]
			ui.Success(fmt.Sprintf("java -version:   %s", firstLine))
		} else {
			ui.Fail(fmt.Sprintf("java -version:   Failed to execute (%v)", err))
		}

		// 2. Java Compiler
		binJavac := filepath.Join(info.Path, "bin", "javac")
		if out, err := exec.Command(binJavac, "-version").CombinedOutput(); err == nil {
			firstLine := strings.Split(string(out), "\n")[0]
			ui.Success(fmt.Sprintf("javac -version:  %s", firstLine))
		} else {
			ui.Fail("javac -version:  Compiler not functional")
		}

		// 3. Gradle Runtime Compatibility
		if _, err := exec.LookPath("gradle"); err == nil {
			cmdG := exec.Command("gradle", "-v")
			cmdG.Env = append(cmdG.Environ(), "JAVA_HOME="+info.Path)
			if out, err := cmdG.CombinedOutput(); err == nil && strings.Contains(string(out), "Gradle") {
				ui.Success(fmt.Sprintf("Gradle Engine:   Compatible (Runs daemon and client on JDK %d)", info.Major))
			} else {
				ui.Warn(fmt.Sprintf("Gradle Engine:   Compatibility notice on JDK %d", info.Major))
			}
		}

		// 4. D8 Dexer Compatibility
		if _, err := exec.LookPath("d8"); err == nil {
			cmdD8 := exec.Command("d8", "--version")
			cmdD8.Env = append(cmdD8.Environ(), "JAVA_HOME="+info.Path)
			if out, err := cmdD8.CombinedOutput(); err == nil {
				ui.Success(fmt.Sprintf("D8 Dexer:        Compatible (%s)", strings.TrimSpace(string(out))))
			} else {
				ui.Warn("D8 Dexer:        Notice on JDK " + fmt.Sprintf("%d", info.Major))
			}
		}

		// 5. Apksigner Compatibility
		if _, err := exec.LookPath("apksigner"); err == nil {
			cmdApk := exec.Command("apksigner", "--version")
			cmdApk.Env = append(cmdApk.Environ(), "JAVA_HOME="+info.Path)
			if out, err := cmdApk.CombinedOutput(); err == nil {
				ui.Success(fmt.Sprintf("apksigner:       Compatible (%s)", strings.TrimSpace(string(out))))
			} else {
				ui.Warn("apksigner:       Notice on JDK " + fmt.Sprintf("%d", info.Major))
			}
		}

		fmt.Println()
	}

	return nil
}
