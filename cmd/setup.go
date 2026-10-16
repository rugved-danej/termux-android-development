// Package cmd implements the user-facing CLI commands for tadev.
package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

// RunSetup guides the user through complete automated setup of the Android development environment.
func RunSetup(force bool, javaProfile string) error {
	ui.PrintBanner()

	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	selectedProfile := cfg.GetConfiguredJavaProfile()

	if javaProfile != "" {
		if p, ok := env.GetJavaProfile(javaProfile); ok {
			selectedProfile = p
		} else {
			ui.Warn(fmt.Sprintf("Unknown Java profile '%s'. Falling back to default %s", javaProfile, selectedProfile.ID))
		}
	} else if ui.IsTerminal() {
		var menuItems []ui.MenuItem
		for _, p := range env.AvailableJavaProfiles {
			menuItems = append(menuItems, ui.MenuItem{
				Key:   p.ID,
				Title: p.Title,
			})
		}

		idx, confirmed, err := ui.RunJavaProfileMenu("SELECT JAVA PROFILE", cfg.DeviceClass, menuItems, env.JavaProfilesDocsURL)
		if err == nil && confirmed && idx >= 0 && idx < len(env.AvailableJavaProfiles) {
			selectedProfile = env.AvailableJavaProfiles[idx]
		}
	}

	totalSteps := "6"
	currentStep := 1

	// Step 1: Detect architecture and environment
	ui.Step(fmt.Sprintf("%d", currentStep), totalSteps, "Detecting System Architecture & Environment")
	currentStep++
	ui.Infof("Device:       %s", ui.Purple(cfg.DeviceClass))
	ui.Infof("Architecture: %s", ui.Purple(cfg.Arch))
	ui.Infof("System RAM:   %.1f GB (Auto-tuned Gradle Heap: %s)", float64(cfg.TotalMemMB)/1024, ui.Purple(cfg.JvmArgs))
	ui.Infof("SDK Location: %s (Hidden)", ui.Purple(cfg.SdkDir))
	ui.Infof("Java Profile: %s", ui.GreenText(selectedProfile.Title))
	ui.Success("Environment validated.")

	// Step 2: Install Termux native packages
	ui.Step(fmt.Sprintf("%d", currentStep), totalSteps, "Installing Termux Native Packages")
	currentStep++
	sp1 := ui.NewSpinner(fmt.Sprintf("Installing %s, AAPT2, Gradle, apksigner, and build tools...", selectedProfile.ID))
	sp1.Start()
	if err := env.InstallTermuxPackagesForProfile(selectedProfile); err != nil {
		sp1.StopFail(err.Error())
		return err
	}
	sp1.StopSuccess("Termux packages installed successfully.")

	// Step 3: Setup latest Android Command-Line Tools
	ui.Step(fmt.Sprintf("%d", currentStep), totalSteps, "Setting Up Latest Android Command-Line Tools")
	currentStep++
	sp2 := ui.NewSpinner("Fetching Command-Line Tools from Google repository...")
	sp2.Start()
	ver, err := env.SetupCmdlineTools(cfg, force, func(downloaded, total int64) {
		if total > 0 {
			mbDown := float64(downloaded) / (1024 * 1024)
			mbTotal := float64(total) / (1024 * 1024)
			bar := ui.RenderProgressBar(downloaded, total, 10)
			sp2.UpdateMessage(fmt.Sprintf("Downloading cmdline-tools: %.1fMB/%.1fMB %s", mbDown, mbTotal, bar))
		}
	})
	if err != nil {
		sp2.StopFail("Failed to install cmdline-tools: " + err.Error())
		return err
	}
	sp2.StopSuccess("Command-Line Tools build " + ver + " configured.")

	// Step 4: Accept licenses and install platforms/build-tools
	ui.Step(fmt.Sprintf("%d", currentStep), totalSteps, "Accepting Licenses & Installing SDK Platforms/Build-Tools")
	currentStep++
	sp3 := ui.NewSpinner("Installing platform-tools, platforms (API 36 & 35), and build-tools 36.0.0...")
	sp3.Start()
	if err := env.InstallSdkPackages(cfg); err != nil {
		sp3.StopFail("SDK installation failed: " + err.Error())
		return err
	}
	sp3.StopSuccess("SDK platforms and build-tools installed successfully.")

	// Step 5: Configure Gradle properties (native AAPT2 override)
	ui.Step(fmt.Sprintf("%d", currentStep), totalSteps, "Configuring Global Gradle Settings (ARM64 AAPT2 Override)")
	currentStep++
	sp4 := ui.NewSpinner(fmt.Sprintf("Setting up native AAPT2 override & toolchain paths (%s)...", selectedProfile.ID))
	sp4.Start()
	if err := env.ApplyJavaProfile(cfg, selectedProfile); err != nil {
		sp4.StopFail("Failed to configure Java profile: " + err.Error())
		return err
	}
	sp4.StopSuccess("Global Gradle properties & Java profile configured.")

	// Step 6: Configure shell RC and global gradlew wrapper
	ui.Step(fmt.Sprintf("%d", currentStep), totalSteps, "Configuring Shell Environment & Global 'gradlew' Wrapper")
	sp5 := ui.NewSpinner("Configuring ~/.zshrc, ~/.bashrc, and installing smart 'gradlew' launcher...")
	sp5.Start()
	if err := env.ConfigureShell(cfg); err != nil {
		sp5.StopFail("Failed to configure shell: " + err.Error())
		return err
	}
	if err := env.InstallGradlewWrapper(cfg); err != nil {
		sp5.StopFail("Failed to install gradlew wrapper: " + err.Error())
		return err
	}
	sp5.StopSuccess("Shell environment and global 'gradlew' wrapper ready.")

	fmt.Println()
	summary := []string{
		fmt.Sprintf("Java Profile:     %s", ui.GreenText(selectedProfile.Title)),
		fmt.Sprintf("SDK Directory:    %s", ui.YellowText(cfg.SdkDir)),
		fmt.Sprintf("Native AAPT2:     %s/aapt2 (%s)", cfg.PrefixBin, cfg.Arch),
		fmt.Sprintf("Global Launcher:  %s", filepath.Join(cfg.PrefixBin, "gradlew")),
	}
	ui.PrintBox("SETUP COMPLETE", summary)

	if ui.IsTerminal() {
		fmt.Printf("\n%s\n", ui.Purple("Clearing console and restarting shell in 2 seconds..."))
		time.Sleep(2 * time.Second)
		return env.ClearAndRestartShell(cfg)
	}

	fmt.Printf("\nRun '%s' to begin building APKs!\n\n", ui.Purple(env.GetShellReloadHint(cfg)))
	return nil
}
