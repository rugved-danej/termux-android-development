package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

const DefaultNdkVersion = "27.2.12479018"

// RunNdkCommand manages Android NDK installations, shebang fixes, and toolchain configurations.
func RunNdkCommand(args []string) error {
	ui.PrintBanner()

	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	subcommand := "status"
	if len(args) > 0 {
		subcommand = strings.ToLower(args[0])
	}

	switch subcommand {
	case "list":
		return runNdkList(cfg)
	case "install", "add", "setup":
		targetVer := DefaultNdkVersion
		if len(args) > 1 {
			targetVer = args[1]
		}
		return runNdkInstall(cfg, targetVer)
	case "status", "show":
		return runNdkStatus(cfg)
	case "remove", "uninstall", "delete":
		targetVer := ""
		if len(args) > 1 {
			targetVer = args[1]
		}
		return runNdkRemove(cfg, targetVer)
	default:
		return fmt.Errorf("unknown ndk action '%s'. Valid actions: status, list, install [version], remove [version]", subcommand)
	}
}

func runNdkList(cfg *env.Config) error {
	versions := env.ListInstalledNdkVersions(cfg)

	fmt.Printf("%sInstalled Android NDK Toolchains:%s\n\n", ui.Bold+ui.Cyan, ui.Reset)
	if len(versions) == 0 {
		ui.Info("No NDK installations detected in " + filepath.Join(cfg.SdkDir, "ndk"))
		ui.Info(fmt.Sprintf("Run '%stadev ndk install%s' to install the recommended LTS version (%s).", ui.Bold+ui.Green, ui.Reset, DefaultNdkVersion))
		return nil
	}

	activeNdk := env.GetActiveNdkPath(cfg)
	for _, v := range versions {
		path := filepath.Join(cfg.SdkDir, "ndk", v)
		if v == "ndk-bundle" {
			path = filepath.Join(cfg.SdkDir, "ndk-bundle")
		}
		activeTag := ""
		if path == activeNdk {
			activeTag = fmt.Sprintf(" %s[ACTIVE]%s", ui.Bold+ui.Green, ui.Reset)
		}
		ui.Success(fmt.Sprintf("NDK %s: %s%s", v, path, activeTag))
	}
	fmt.Println()
	return nil
}

func runNdkStatus(cfg *env.Config) error {
	fmt.Printf("%sAndroid NDK Environment Status%s\n\n", ui.Bold+ui.Cyan, ui.Reset)

	activePath := env.GetActiveNdkPath(cfg)
	if activePath == "" {
		ui.Warn("No Android NDK installed.")
		ui.Info(fmt.Sprintf("Run '%stadev ndk install%s' to download and configure NDK %s.", ui.Bold+ui.Green, ui.Reset, DefaultNdkVersion))
		return nil
	}

	ui.Success(fmt.Sprintf("Active NDK Path: %s", activePath))

	ndkBuild := filepath.Join(activePath, "ndk-build")
	if info, err := os.Stat(ndkBuild); err == nil && (info.Mode()&0111 != 0) {
		ui.Success("ndk-build executable: Available and executable")
	} else {
		ui.Warn("ndk-build: Missing or not executable (fixing shebangs...)")
		env.FixNdkShebangs(cfg, activePath)
	}

	envNdk := os.Getenv("ANDROID_NDK_HOME")
	if envNdk != "" {
		ui.Success(fmt.Sprintf("ANDROID_NDK_HOME in current shell: %s", envNdk))
	} else {
		ui.Info("ANDROID_NDK_HOME not yet loaded in current terminal.")
		ui.Info(fmt.Sprintf("Run '%s' to apply shell variables.", env.GetShellReloadHint(cfg)))
	}
	fmt.Println()
	return nil
}

func runNdkInstall(cfg *env.Config, version string) error {
	normalizedPkg := version
	if !strings.HasPrefix(normalizedPkg, "ndk;") && normalizedPkg != "ndk-bundle" {
		normalizedPkg = "ndk;" + normalizedPkg
	}

	sp := ui.NewSpinner(fmt.Sprintf("Downloading and installing Android NDK (%s)...", normalizedPkg))
	sp.Start()

	if err := env.InstallSdkPackages(cfg, normalizedPkg); err != nil {
		sp.StopFail("NDK installation failed: " + err.Error())
		return err
	}

	sp.UpdateMessage("Configuring NDK toolchains and fixing Termux shebangs...")
	activeNdk := env.GetActiveNdkPath(cfg)
	if activeNdk != "" {
		env.FixNdkShebangs(cfg, activeNdk)
	}

	_ = env.ConfigureShell(cfg)

	sp.StopSuccess(fmt.Sprintf("Android NDK (%s) installed and configured successfully!", normalizedPkg))

	fmt.Println()
	summary := []string{
		fmt.Sprintf("NDK Location:     %s", activeNdk),
		fmt.Sprintf("Target Package:   %s", normalizedPkg),
		fmt.Sprintf("Shebangs:         Fixed for Termux & Linux native execution"),
		fmt.Sprintf("Shell Variable:   ANDROID_NDK_HOME configured"),
	}
	ui.PrintBox("NDK INSTALLATION COMPLETE", summary)

	fmt.Printf("\n%sRun %s'%s'%s%s to load NDK into your current shell session.%s\n\n",
		ui.Cyan, ui.Bold+ui.White, env.GetShellReloadHint(cfg), ui.Reset, ui.Cyan, ui.Reset)
	return nil
}

func runNdkRemove(cfg *env.Config, version string) error {
	activePath := env.GetActiveNdkPath(cfg)
	if version != "" {
		activePath = filepath.Join(cfg.SdkDir, "ndk", version)
	}
	if activePath == "" || !dirExists(activePath) {
		return fmt.Errorf("no NDK found to remove at %s", activePath)
	}

	sp := ui.NewSpinner(fmt.Sprintf("Removing NDK at %s...", activePath))
	sp.Start()

	if err := os.RemoveAll(activePath); err != nil {
		sp.StopFail("Failed to remove NDK: " + err.Error())
		return err
	}

	_ = env.ConfigureShell(cfg)
	sp.StopSuccess("NDK removed and shell configurations updated.")
	return nil
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
