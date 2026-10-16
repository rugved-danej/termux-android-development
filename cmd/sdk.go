package cmd

import (
	"fmt"
	"strings"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

// RunSdkManager provides granular inspection and package installation via sdkmanager.
func RunSdkManager(args []string) error {
	ui.PrintBanner()

	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	subcommand := "list"
	if len(args) > 0 {
		subcommand = strings.ToLower(args[0])
	}

	switch subcommand {
	case "list", "installed":
		platforms, buildTools := env.ListInstalledSdkPackages(cfg)

		fmt.Printf("%s\n", ui.BrightPurple("◆ Installed SDK Platforms"))
		if len(platforms) == 0 {
			ui.Info("No platforms installed yet.")
		} else {
			for _, p := range platforms {
				ui.Success(p)
			}
		}

		fmt.Printf("\n%s\n", ui.BrightPurple("◆ Installed Build-Tools"))
		if len(buildTools) == 0 {
			ui.Info("No build-tools installed yet.")
		} else {
			for _, bt := range buildTools {
				ui.Success(bt)
			}
		}
		fmt.Println()
		return nil

	case "install", "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: tadev sdk install <package-name>\nExample: tadev sdk install \"platforms;android-34\"")
		}
		pkgs := args[1:]
		sp := ui.NewSpinner(fmt.Sprintf("Installing SDK packages (%s)...", strings.Join(pkgs, ", ")))
		sp.Start()

		if err := env.InstallSdkPackages(cfg, pkgs...); err != nil {
			sp.StopFail(err.Error())
			return err
		}
		sp.StopSuccess("SDK packages installed successfully.")
		return nil

	default:
		return fmt.Errorf("unknown sdk action '%s'. Valid actions: list, install <pkg>", subcommand)
	}
}
