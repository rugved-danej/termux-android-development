package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

type UninstallFlags struct {
	AutoConfirm   bool
	All           bool
	Self          bool
	SdkOnly       bool
	PurgePackages bool
}

// RunUninstall guides or executes uninstallation of selected Android development components.
func RunUninstall(flags UninstallFlags) error {
	ui.PrintBanner()

	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	selected := map[string]bool{
		"sdk":         true,
		"shell":       true,
		"gradleprops": true,
		"wrapper":     true,
		"config":      true,
		"jdk26":       false,
		"jdk25":       false,
		"jdk21":       false,
		"jdk17":       false,
		"gradle":      false,
		"tools":       false,
		"ndk":         false,
		"self":        false,
	}

	if flags.All {
		for k := range selected {
			selected[k] = true
		}
	} else if flags.Self {
		for k := range selected {
			selected[k] = false
		}
		selected["self"] = true
		selected["config"] = true
	} else if flags.SdkOnly {
		for k := range selected {
			selected[k] = false
		}
		selected["sdk"] = true
	} else if flags.PurgePackages {
		selected["jdk26"] = true
		selected["jdk25"] = true
		selected["jdk21"] = true
		selected["jdk17"] = true
		selected["gradle"] = true
		selected["tools"] = true
	}

	// Interactive TUI Selection if terminal is attached and not auto-confirmed
	if !flags.AutoConfirm && ui.IsTerminal() && !flags.All && !flags.Self && !flags.SdkOnly {
		items := []*ui.CheckboxOption{
			{ID: "sdk", Label: "Android SDK Directory", Description: cfg.SdkDir + " (Platforms, build-tools, cmdline-tools)", Checked: selected["sdk"]},
			{ID: "shell", Label: "Shell Configurations", Description: "Exports & aliases in ~/.bashrc, ~/.zshrc, fish, and profile configs", Checked: selected["shell"]},
			{ID: "gradleprops", Label: "Gradle Properties", Description: "AAPT2 override & toolchain paths in ~/.gradle/gradle.properties", Checked: selected["gradleprops"]},
			{ID: "wrapper", Label: "Global 'gradlew' Wrapper", Description: cfg.GradlewWrapper, Checked: selected["wrapper"]},
			{ID: "jdk26", Label: "OpenJDK 26 Runtime (TUR)", Description: "Package: openjdk-26 (via " + cfg.PkgManager + ")", Checked: selected["jdk26"]},
			{ID: "jdk25", Label: "OpenJDK 25 Runtime", Description: "Package: openjdk-25 (via " + cfg.PkgManager + ")", Checked: selected["jdk25"]},
			{ID: "jdk21", Label: "OpenJDK 21 Toolchains", Description: "Package: openjdk-21 (via " + cfg.PkgManager + ")", Checked: selected["jdk21"]},
			{ID: "jdk17", Label: "OpenJDK 17 LTS Runtime", Description: "Package: openjdk-17 (via " + cfg.PkgManager + ")", Checked: selected["jdk17"]},
			{ID: "gradle", Label: "Gradle Build System", Description: "Package: gradle (via " + cfg.PkgManager + ")", Checked: selected["gradle"]},
			{ID: "tools", Label: "Android Tools & AAPT2", Description: "Packages: aapt2, apksigner, d8, android-tools", Checked: selected["tools"]},
			{ID: "ndk", Label: "Android NDK Toolchains", Description: filepath.Join(cfg.SdkDir, "ndk") + " (C/C++ compilers)", Checked: selected["ndk"]},
			{ID: "config", Label: "tadev Configuration & Cache", Description: cfg.ConfigDir + " and update caches", Checked: selected["config"]},
			{ID: "self", Label: "Self-Remove tadev CLI Tool", Description: "Completely delete " + filepath.Join(cfg.PrefixBin, "tadev") + " from device", Checked: selected["self"]},
		}

		updated, confirmed, err := ui.RunCheckboxMenu("UNINSTALLATION CHOOSER", "Select components to remove from your system:", items)
		if err != nil || !confirmed {
			ui.Info("Uninstallation cancelled by user.")
			return nil
		}

		for _, it := range updated {
			selected[it.ID] = it.Checked
		}
	} else if !flags.AutoConfirm {
		// Non-interactive fallback prompt
		fmt.Printf("%s%sProceed with uninstallation? (y/N): %s", ui.Yellow, ui.Bold, ui.Reset)
		reader := bufio.NewReader(os.Stdin)
		res, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(res)) != "y" && strings.TrimSpace(strings.ToLower(res)) != "yes" {
			ui.Info("Uninstallation cancelled by user.")
			return nil
		}
	}

	anySelected := false
	for _, v := range selected {
		if v {
			anySelected = true
			break
		}
	}
	if !anySelected {
		ui.Info("No components selected for uninstallation.")
		return nil
	}

	fmt.Println()

	// 1. Remove SDK
	if selected["sdk"] {
		sp := ui.NewSpinner("Removing Android SDK directory (" + cfg.SdkDir + ")...")
		sp.Start()
		if err := os.RemoveAll(cfg.SdkDir); err != nil {
			sp.StopFail("Failed to remove SDK: " + err.Error())
		} else {
			sp.StopSuccess("Android SDK removed.")
		}
	} else if selected["ndk"] {
		sp := ui.NewSpinner("Removing Android NDK toolchains...")
		sp.Start()
		_ = os.RemoveAll(filepath.Join(cfg.SdkDir, "ndk"))
		_ = os.RemoveAll(filepath.Join(cfg.SdkDir, "ndk-bundle"))
		sp.StopSuccess("Android NDK removed.")
	}

	// 2. Clean shell configs
	if selected["shell"] {
		sp := ui.NewSpinner("Restoring shell configurations to clean state...")
		sp.Start()
		if err := env.CleanShell(cfg); err != nil {
			sp.StopFail("Failed to clean shell: " + err.Error())
		} else {
			sp.StopSuccess("Shell configurations restored.")
		}
	}

	// 3. Clean gradle properties
	if selected["gradleprops"] {
		sp := ui.NewSpinner("Restoring ~/.gradle/gradle.properties...")
		sp.Start()
		if err := env.CleanGradleProperties(cfg); err != nil {
			sp.StopFail("Failed to clean gradle.properties: " + err.Error())
		} else {
			sp.StopSuccess("Gradle properties restored.")
		}
	}

	// 4. Remove gradlew wrapper
	if selected["wrapper"] {
		sp := ui.NewSpinner("Removing global 'gradlew' wrapper...")
		sp.Start()
		if err := env.RemoveGradlewWrapper(cfg); err != nil {
			sp.StopFail("Failed to remove wrapper: " + err.Error())
		} else {
			sp.StopSuccess("Global 'gradlew' wrapper removed.")
		}
	}

	// 5. Purge Packages
	var pkgsToPurge []string
	if selected["jdk26"] {
		pkgsToPurge = append(pkgsToPurge, "openjdk-26")
	}
	if selected["jdk25"] {
		pkgsToPurge = append(pkgsToPurge, "openjdk-25")
	}
	if selected["jdk21"] {
		pkgsToPurge = append(pkgsToPurge, "openjdk-21")
	}
	if selected["jdk17"] {
		pkgsToPurge = append(pkgsToPurge, "openjdk-17")
	}
	if selected["gradle"] {
		pkgsToPurge = append(pkgsToPurge, "gradle")
	}
	if selected["tools"] {
		pkgsToPurge = append(pkgsToPurge, "aapt2", "apksigner", "d8", "android-tools")
	}

	if len(pkgsToPurge) > 0 {
		sp := ui.NewSpinner(fmt.Sprintf("Uninstalling packages (%s)...", strings.Join(pkgsToPurge, ", ")))
		sp.Start()
		if err := env.UninstallPackages(pkgsToPurge); err != nil {
			sp.StopFail("Failed to remove packages: " + err.Error())
		} else {
			sp.StopSuccess("Selected packages uninstalled.")
		}
	}

	// 6. Clean config & cache
	if selected["config"] {
		_ = os.RemoveAll(cfg.ConfigDir)
		_ = os.Remove(cfg.UpdateCache)
		ui.Success("Configuration and cache cleared.")
	}

	// 7. Self-removal of tadev binary
	if selected["self"] {
		sp := ui.NewSpinner("Removing tadev and termux-android-dev binaries from system...")
		sp.Start()
		tadevBin := filepath.Join(cfg.PrefixBin, "tadev")
		altBin := filepath.Join(cfg.PrefixBin, "termux-android-dev")
		_ = os.Remove(altBin)
		_ = os.Remove(tadevBin)
		sp.StopSuccess("tadev CLI binaries deleted.")
		fmt.Printf("\n%s[OK] Complete uninstallation finished! tadev has been completely removed from your device.%s\n\n", ui.Green+ui.Bold, ui.Reset)
		return nil
	}

	fmt.Printf("\n%s[OK] Selected components successfully uninstalled!%s\n", ui.Green+ui.Bold, ui.Reset)
	fmt.Printf("%sRestart your shell or run %s'%s'%s%s to finalize changes.%s\n\n",
		ui.Cyan, ui.White+ui.Bold, env.GetShellReloadHint(cfg), ui.Reset, ui.Cyan, ui.Reset)

	return nil
}
