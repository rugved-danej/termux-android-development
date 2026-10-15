package env

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// JavaProfile represents a configurable Java runtime & Gradle toolchain profile.
type JavaProfile struct {
	ID            string   // Canonical identifier: "25+21", "21", "26+21", "26", "17", "25+17"
	Title         string   // Human-readable menu title
	Description   string   // Detailed explanation for Ctrl+D inspection
	Packages      []string // OS packages required (e.g. "openjdk-25", "openjdk-21")
	HostJavaDir   string   // Primary host JVM dirname inside jvmDir (e.g. "java-25-openjdk")
	ToolchainDirs []string // JVM dirnames registered in gradle.properties toolchain paths
	RequiresTur   bool     // True if package requires tur-repo (e.g. openjdk-26)
	IsDefault     bool     // True for default recommended profile
}

// JavaProfilesDocsURL is the documentation URL detailing architecture differences between combinations and solos.
const JavaProfilesDocsURL = "https://tadev.vercel.app/docs/java-profiles"

// AvailableJavaProfiles defines the supported Java profiles using exclusively official Termux stable packages.
var AvailableJavaProfiles = []JavaProfile{
	// --- Recommended Single-JDK Solo (Best for Termux Android Development) ---
	{
		ID:          "21",
		Title:       "[21]      OpenJDK 21 LTS Only (Single JDK, 100% AGP Stable, Saves ~300MB) [Recommended]",
		Description: "Single OpenJDK 21 LTS runtime used for both Termux host and Android Gradle builds. Saves ~300MB disk space, zero bytecode conflicts, perfectly stable with all Android Gradle Plugin (AGP 8.x/9.x) versions.",
		Packages:    []string{"openjdk-21"},
		HostJavaDir: "java-21-openjdk",
		ToolchainDirs: []string{
			"java-21-openjdk",
		},
		RequiresTur: false,
		IsDefault:   true,
	},

	// --- Dual-JDK Combinations ---
	{
		ID:          "25+21",
		Title:       "[25 + 21] OpenJDK 25 (Host) + OpenJDK 21 (Gradle Toolchain)",
		Description: "Modern OpenJDK 25 host paired with stable OpenJDK 21 LTS Gradle toolchain. Best for modern Termux host performance with 100% stable Android Gradle builds.",
		Packages:    []string{"openjdk-25", "openjdk-21"},
		HostJavaDir: "java-25-openjdk",
		ToolchainDirs: []string{
			"java-25-openjdk",
			"java-21-openjdk",
		},
		RequiresTur: false,
		IsDefault:   false,
	},
	{
		ID:          "25+17",
		Title:       "[25 + 17] OpenJDK 25 (Host) + OpenJDK 17 (Gradle Toolchain)",
		Description: "Modern OpenJDK 25 for Termux host commands paired with OpenJDK 17 LTS toolchain for older Android projects that require Java 17 toolchain compatibility.",
		Packages:    []string{"openjdk-25", "openjdk-17"},
		HostJavaDir: "java-25-openjdk",
		ToolchainDirs: []string{
			"java-25-openjdk",
			"java-17-openjdk",
		},
		RequiresTur: false,
		IsDefault:   false,
	},
	{
		ID:          "21+17",
		Title:       "[21 + 17] OpenJDK 21 (Host) + OpenJDK 17 (Gradle Toolchain)",
		Description: "LTS OpenJDK 21 host paired with legacy OpenJDK 17 LTS toolchain for older Android projects.",
		Packages:    []string{"openjdk-21", "openjdk-17"},
		HostJavaDir: "java-21-openjdk",
		ToolchainDirs: []string{
			"java-21-openjdk",
			"java-17-openjdk",
		},
		RequiresTur: false,
		IsDefault:   false,
	},

	// --- Other Single-JDK Solos ---
	{
		ID:          "25",
		Title:       "[25]      OpenJDK 25 LTS Only (Single JDK, Modern 2025 LTS Standard)",
		Description: "Single modern OpenJDK 25 LTS runtime for host and Gradle. Maximum runtime optimization and language features for Gradle 9+ and modern Android builds.",
		Packages:    []string{"openjdk-25"},
		HostJavaDir: "java-25-openjdk",
		ToolchainDirs: []string{
			"java-25-openjdk",
		},
		RequiresTur: false,
		IsDefault:   false,
	},
	{
		ID:          "17",
		Title:       "[17]      OpenJDK 17 LTS Only (Single JDK, AGP 8.0 Baseline, Saves ~330MB)",
		Description: "Single OpenJDK 17 LTS runtime for host and Gradle. Maximum legacy compatibility for older Android projects, legacy Gradle plugins, and Android Gradle Plugin 8.0 baseline requirements.",
		Packages:    []string{"openjdk-17"},
		HostJavaDir: "java-17-openjdk",
		ToolchainDirs: []string{
			"java-17-openjdk",
		},
		RequiresTur: false,
		IsDefault:   false,
	},
}

// NormalizeJavaProfileID converts loose user input into canonical profile ID.
func NormalizeJavaProfileID(input string) string {
	cleaned := strings.ToLower(strings.TrimSpace(input))
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "+")
	cleaned = strings.ReplaceAll(cleaned, "&", "+")
	cleaned = strings.ReplaceAll(cleaned, "and", "+")
	cleaned = strings.ReplaceAll(cleaned, "_", "+")
	cleaned = strings.ReplaceAll(cleaned, "-", "+")

	switch cleaned {
	case "default":
		return "21"
	case "25+21":
		return "25+21"
	case "25+17":
		return "25+17"
	case "21+17":
		return "21+17"
	case "25", "jdk25", "openjdk25", "25lts":
		return "25"
	case "21", "jdk21", "openjdk21", "21lts":
		return "21"
	case "17", "jdk17", "openjdk17", "17lts":
		return "17"
	default:
		return cleaned
	}
}

// GetJavaProfile returns the profile by canonical ID or normalized input.
func GetJavaProfile(id string) (JavaProfile, bool) {
	norm := NormalizeJavaProfileID(id)
	for _, p := range AvailableJavaProfiles {
		if p.ID == norm {
			return p, true
		}
	}
	// Fallback check on exact ID
	for _, p := range AvailableJavaProfiles {
		if strings.EqualFold(p.ID, id) {
			return p, true
		}
	}
	return JavaProfile{}, false
}

// DefaultJavaProfile returns the canonical recommended default.
func DefaultJavaProfile() JavaProfile {
	for _, p := range AvailableJavaProfiles {
		if p.IsDefault {
			return p
		}
	}
	return AvailableJavaProfiles[0]
}

// EnsureTurRepoInstalled ensures the Termux User Repository (tur-repo) is installed.
func EnsureTurRepoInstalled() error {
	if IsPackageInstalled("tur-repo") {
		return nil
	}

	pm := DetectPackageManager()
	var args []string
	switch pm {
	case "pacman":
		args = []string{"-S", "--noconfirm", "tur-repo"}
	default:
		args = []string{"install", "-y", "tur-repo"}
	}

	cmd := exec.Command(pm, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install tur-repo: %v\nOutput: %s", err, out.String())
	}

	// Update package lists to make TUR packages available
	updateCmd := exec.Command(pm, "update", "-y")
	_ = updateCmd.Run()

	return nil
}

// SetSystemDefaultJava updates Termux's update-alternatives to point to the desired Java binary.
func SetSystemDefaultJava(cfg *Config, hostDir string) error {
	javaBin := filepath.Join(cfg.JvmDir, hostDir, "bin", "java")
	if _, err := os.Stat(javaBin); err != nil {
		return fmt.Errorf("java binary not found at %s", javaBin)
	}

	if _, err := exec.LookPath("update-alternatives"); err == nil {
		cmd := exec.Command("update-alternatives", "--set", "java", javaBin)
		_ = cmd.Run()
	}

	return nil
}

// ApplyJavaProfile applies the chosen Java profile across configuration, Gradle, and shell.
func ApplyJavaProfile(cfg *Config, profile JavaProfile) error {
	// 1. If tur-repo is required, ensure it is installed
	if profile.RequiresTur && cfg.IsTermux {
		if err := EnsureTurRepoInstalled(); err != nil {
			return err
		}
	}

	// 2. Install missing packages for profile
	var missing []string
	for _, pkg := range profile.Packages {
		if !IsPackageInstalled(pkg) {
			missing = append(missing, pkg)
		}
	}
	if len(missing) > 0 {
		pm := DetectPackageManager()
		var args []string
		switch pm {
		case "pacman":
			args = append([]string{"-S", "--noconfirm"}, missing...)
		default:
			args = append([]string{"install", "-y"}, missing...)
		}
		cmd := exec.Command(pm, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install profile packages %v: %v\nOutput: %s", missing, err, out.String())
		}
	}

	// 3. Set update-alternatives if possible
	_ = SetSystemDefaultJava(cfg, profile.HostJavaDir)

	// 4. Update ~/.gradle/gradle.properties toolchains
	var toolchainPaths []string
	for _, dir := range profile.ToolchainDirs {
		toolchainPaths = append(toolchainPaths, filepath.Join(cfg.JvmDir, dir))
	}
	if err := ConfigureGradlePropertiesWithPaths(cfg, toolchainPaths); err != nil {
		return err
	}

	// 5. Update Shell files
	if err := ConfigureShellWithJava(cfg, filepath.Join(cfg.JvmDir, profile.HostJavaDir)); err != nil {
		return err
	}

	// 6. Save preference to config file
	return cfg.SaveJavaProfileSetting(profile.ID)
}

// InstalledJavaInfo contains diagnostic information about an installed JDK.
type InstalledJavaInfo struct {
	DirName     string
	Path        string
	VersionLine string
	Major       int
	IsActive    bool
}

// DetectInstalledJavaVersions scans the system's JVM directory for installed OpenJDKs.
func DetectInstalledJavaVersions(cfg *Config) []InstalledJavaInfo {
	var list []InstalledJavaInfo
	candidates := []struct {
		dir   string
		major int
	}{
		{"java-26-openjdk", 26},
		{"java-25-openjdk", 25},
		{"java-21-openjdk", 21},
		{"java-17-openjdk", 17},
		{"java-11-openjdk", 11},
	}

	activeJavaPath := ""
	if p, err := exec.LookPath("java"); err == nil {
		if rp, err := filepath.EvalSymlinks(p); err == nil {
			activeJavaPath = rp
		}
	}

	for _, c := range candidates {
		jvmPath := filepath.Join(cfg.JvmDir, c.dir)
		binJava := filepath.Join(jvmPath, "bin", "java")
		if _, err := os.Stat(binJava); err == nil {
			verLine := "Unknown"
			if out, err := exec.Command(binJava, "-version").CombinedOutput(); err == nil {
				lines := strings.Split(string(out), "\n")
				if len(lines) > 0 {
					verLine = strings.TrimSpace(lines[0])
				}
			}

			isActive := false
			if activeJavaPath != "" && strings.HasPrefix(activeJavaPath, jvmPath) {
				isActive = true
			}

			list = append(list, InstalledJavaInfo{
				DirName:     c.dir,
				Path:        jvmPath,
				VersionLine: verLine,
				Major:       c.major,
				IsActive:    isActive,
			})
		}
	}

	return list
}
