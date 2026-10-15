package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

func RunDoctor() error {
	ui.PrintBanner()

	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	fmt.Printf("%s%sSystem Diagnostic & Environment Health Check%s\n\n", ui.Bold, ui.Cyan, ui.Reset)

	issues := 0

	// 1. Device Class & Architecture
	ui.Success(fmt.Sprintf("Device: %s%s%s", ui.Bold, cfg.DeviceClass, ui.Reset))
	ui.Success(fmt.Sprintf("Architecture: %s%s%s", ui.Bold, cfg.Arch, ui.Reset))
	ui.Success(fmt.Sprintf("Memory: %s%.1f GB physical RAM%s (Auto-tuned JVM args: %s%s%s)", ui.Bold, float64(cfg.TotalMemMB)/1024, ui.Reset, ui.Cyan, cfg.JvmArgs, ui.Reset))
	ui.Success(fmt.Sprintf("Environment: Prefix %s (Package Manager: %s)", cfg.Prefix, cfg.PkgManager))

	// 2. Configured Java Profile
	prof := cfg.GetConfiguredJavaProfile()
	ui.Success(fmt.Sprintf("Configured Java Profile: %s%s%s (%s)", ui.Bold, prof.ID, ui.Reset, prof.Title))

	// Check required packages for active profile
	for _, pkg := range prof.Packages {
		if !env.IsPackageInstalled(pkg) {
			ui.Fail(fmt.Sprintf("Profile Package '%s': Not installed (run 'tadev java switch %s')", pkg, prof.ID))
			issues++
		}
	}

	// 3. Detected OpenJDK Runtimes
	installedJDKs := env.DetectInstalledJavaVersions(cfg)
	if len(installedJDKs) == 0 {
		ui.Fail("No OpenJDK installations detected in " + cfg.JvmDir)
		issues++
	} else {
		for _, info := range installedJDKs {
			role := "Installed"
			if info.DirName == prof.HostJavaDir {
				role = "Host / PATH"
			} else {
				for _, tc := range prof.ToolchainDirs {
					if info.DirName == tc {
						role = "Toolchain"
						break
					}
				}
			}
			ui.Success(fmt.Sprintf("OpenJDK %d [%s]: %s (%s)", info.Major, role, info.VersionLine, info.Path))
		}
	}

	// 4. Active Java in PATH
	if p, err := exec.LookPath("java"); err == nil {
		ui.Success(fmt.Sprintf("Active Java command in PATH: %s", p))
	} else {
		ui.Fail("No 'java' command found in PATH")
		issues++
	}

	// 5. Android SDK Directory Check
	if info, err := os.Stat(cfg.SdkDir); err == nil && info.IsDir() {
		ui.Success(fmt.Sprintf("Android SDK Directory: %s (Hidden)", cfg.SdkDir))
	} else {
		ui.Fail(fmt.Sprintf("Android SDK Directory not found at %s", cfg.SdkDir))
		issues++
	}

	// 6. Command-Line Tools Check
	sdkmanager := filepath.Join(cfg.CmdlineDir, "bin", "sdkmanager")
	if info, err := os.Stat(sdkmanager); err == nil && (info.Mode()&0111 != 0) {
		ui.Success(fmt.Sprintf("Command-Line Tools: Available (%s)", cfg.CmdlineDir))
	} else {
		ui.Fail(fmt.Sprintf("sdkmanager not found or not executable at %s", sdkmanager))
		issues++
	}

	// 7. Native AAPT2 Binary Check
	aapt2 := filepath.Join(cfg.PrefixBin, "aapt2")
	if out, err := exec.Command(aapt2, "version").CombinedOutput(); err == nil {
		firstLine := strings.Split(string(out), "\n")[0]
		ui.Success(fmt.Sprintf("Native AAPT2 Binary: %s (%s)", aapt2, firstLine))
	} else {
		ui.Fail(fmt.Sprintf("Native AAPT2 binary not working at %s", aapt2))
		issues++
	}

	// 8. Gradle Properties Check
	if data, err := os.ReadFile(cfg.GradleProps); err == nil {
		if strings.Contains(string(data), "android.aapt2FromMavenOverride") {
			ui.Success("Gradle Properties: AAPT2 native override properly configured")
		} else {
			ui.Fail("Gradle Properties: android.aapt2FromMavenOverride missing in " + cfg.GradleProps)
			issues++
		}
	} else {
		ui.Fail("Gradle Properties file not found at " + cfg.GradleProps)
		issues++
	}

	// 9. Global gradlew Wrapper Check
	if info, err := os.Stat(cfg.GradlewWrapper); err == nil && (info.Mode()&0111 != 0) {
		ui.Success(fmt.Sprintf("Global 'gradlew' wrapper: Installed at %s", cfg.GradlewWrapper))
	} else {
		ui.Warn(fmt.Sprintf("Global 'gradlew' wrapper missing at %s", cfg.GradlewWrapper))
	}

	// 10. Installed Platforms & Build-Tools
	platsDir := filepath.Join(cfg.SdkDir, "platforms")
	if entries, err := os.ReadDir(platsDir); err == nil {
		var names []string
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		ui.Success(fmt.Sprintf("Installed Platforms: %s", strings.Join(names, ", ")))
	}

	btDir := filepath.Join(cfg.SdkDir, "build-tools")
	if entries, err := os.ReadDir(btDir); err == nil {
		var names []string
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		ui.Success(fmt.Sprintf("Installed Build-Tools: %s", strings.Join(names, ", ")))
	}

	// 11. Android NDK Check
	ndkVersions := env.ListInstalledNdkVersions(cfg)
	if len(ndkVersions) > 0 {
		ui.Success(fmt.Sprintf("Android NDK: Installed (%s)", strings.Join(ndkVersions, ", ")))
		if ndkHome := env.GetActiveNdkPath(cfg); ndkHome != "" {
			ui.Success(fmt.Sprintf("Active NDK Path: %s", ndkHome))
		}
	} else {
		ui.Info("Android NDK: Not installed (optional, run 'tadev ndk install' to set up)")
	}

	// 12. Shell Configurations Check
	shellFiles := env.GetShellConfigFiles(cfg)
	var activeConfigs []string
	for _, sf := range shellFiles {
		if data, err := os.ReadFile(sf); err == nil && strings.Contains(string(data), cfg.BlockStart) {
			activeConfigs = append(activeConfigs, filepath.Base(sf))
		}
	}
	if len(activeConfigs) > 0 {
		ui.Success(fmt.Sprintf("Shell Configurations: Active in %s", strings.Join(activeConfigs, ", ")))
	} else {
		ui.Warn(fmt.Sprintf("Shell Configurations: Environment variables not yet configured in shells (run 'tadev install' or 'tadev java switch %s')", prof.ID))
	}

	fmt.Println()
	if issues == 0 {
		ui.Success("Your Android development environment is 100% healthy and ready to build APKs!")
	} else {
		ui.Warn(fmt.Sprintf("Detected %d issue(s). Run 'tadev setup' to resolve them automatically.", issues))
	}
	fmt.Println()

	return nil
}
