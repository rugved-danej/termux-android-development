package env

import (
	"bytes"
	"fmt"
	"os/exec"
)

var BasePackages = []string{
	"gradle",
	"android-tools",
	"aapt2",
	"apksigner",
	"d8",
	"wget",
	"unzip",
	"curl",
	"git",
	"file",
}

// RequiredPackages is maintained for backwards compatibility.
var RequiredPackages = append([]string{"openjdk-25", "openjdk-21"}, BasePackages...)

// GetPackagesForProfile returns the complete list of packages needed for a Java profile.
func GetPackagesForProfile(profile JavaProfile) []string {
	var pkgs []string
	for _, p := range profile.Packages {
		if p != "tur-repo" {
			pkgs = append(pkgs, p)
		}
	}
	pkgs = append(pkgs, BasePackages...)
	return pkgs
}

// InstallTermuxPackages installs default packages using the detected package manager.
func InstallTermuxPackages() error {
	return InstallTermuxPackagesForProfile(DefaultJavaProfile())
}

// InstallTermuxPackagesForProfile installs core packages for a specific Java profile.
func InstallTermuxPackagesForProfile(profile JavaProfile) error {
	if profile.RequiresTur {
		if err := EnsureTurRepoInstalled(); err != nil {
			return err
		}
	}

	pkgs := GetPackagesForProfile(profile)
	pm := DetectPackageManager()
	var args []string
	switch pm {
	case "pacman":
		args = append([]string{"-S", "--noconfirm"}, pkgs...)
	default:
		args = append([]string{"install", "-y"}, pkgs...)
	}

	cmd := exec.Command(pm, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s install failed: %v\nOutput: %s", pm, err, out.String())
	}
	return nil
}

// IsPackageInstalled checks if a given package is installed on the host.
func IsPackageInstalled(pkg string) bool {
	if _, err := exec.LookPath("dpkg"); err == nil {
		cmd := exec.Command("dpkg", "-s", pkg)
		return cmd.Run() == nil
	}
	_, err := exec.LookPath(pkg)
	return err == nil
}

// UninstallPackages removes specified packages using the native package manager.
func UninstallPackages(pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}

	var toRemove []string
	for _, p := range pkgs {
		if IsPackageInstalled(p) {
			toRemove = append(toRemove, p)
		}
		if p == "openjdk-26" && IsPackageInstalled("openjdk-26-x") {
			toRemove = append(toRemove, "openjdk-26-x")
		}
		if p == "openjdk-25" && IsPackageInstalled("openjdk-25-x") {
			toRemove = append(toRemove, "openjdk-25-x")
		}
		if p == "openjdk-21" && IsPackageInstalled("openjdk-21-x") {
			toRemove = append(toRemove, "openjdk-21-x")
		}
		if p == "openjdk-17" && IsPackageInstalled("openjdk-17-x") {
			toRemove = append(toRemove, "openjdk-17-x")
		}
	}

	if len(toRemove) == 0 {
		return nil
	}

	pm := DetectPackageManager()
	var args []string
	switch pm {
	case "pkg":
		args = append([]string{"uninstall", "-y"}, toRemove...)
	case "apt", "apt-get":
		args = append([]string{"purge", "-y"}, toRemove...)
	case "pacman":
		args = append([]string{"-R", "--noconfirm"}, toRemove...)
	case "dnf", "zypper":
		args = append([]string{"remove", "-y"}, toRemove...)
	default:
		args = append([]string{"remove", "-y"}, toRemove...)
	}

	cmd := exec.Command(pm, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s uninstall failed: %v\nOutput: %s", pm, err, out.String())
	}
	return nil
}

// CheckPackageUpdates checks for package updates across repositories.
func CheckPackageUpdates() ([]string, error) {
	pm := DetectPackageManager()
	var cmd *exec.Cmd

	switch pm {
	case "pkg":
		cmd = exec.Command("pkg", "list-upgradable")
	case "apt", "apt-get":
		cmd = exec.Command("apt", "list", "--upgradable")
	default:
		return nil, nil
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	var updates []string
	lines := bytes.Split(out, []byte("\n"))
	for _, l := range lines {
		line := string(l)
		for _, pkg := range []string{"openjdk", "gradle", "android-tools", "aapt2"} {
			if bytes.Contains([]byte(line), []byte(pkg)) {
				updates = append(updates, line)
				break
			}
		}
	}
	return updates, nil
}
