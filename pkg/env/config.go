package env

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	HomeDir        string
	AndroidDir     string
	SdkDir         string
	CmdlineDir     string
	Prefix         string
	PrefixBin      string
	JvmDir         string
	GradleDir      string
	GradleProps    string
	GradlewWrapper string
	UpdateCache    string
	ConfigDir      string
	ConfigFile     string
	BlockStart     string
	BlockEnd       string
	Arch           string
	DeviceClass    string
	IsTermux       bool
	PkgManager     string
	TotalMemMB     int
	OptimalJvmArgs string
	JvmArgs        string
	JavaProfile    string
}

// DetectPrefix dynamically determines the prefix across Termux, standard Linux, Chromebooks, and WSA.
func DetectPrefix() (string, bool) {
	if p := os.Getenv("PREFIX"); p != "" {
		isTermux := strings.Contains(p, "com.termux")
		return p, isTermux
	}
	if _, err := os.Stat("/data/data/com.termux/files/usr"); err == nil {
		return "/data/data/com.termux/files/usr", true
	}
	if _, err := os.Stat("/usr/local/bin"); err == nil {
		return "/usr/local", false
	}
	return "/usr", false
}

// DetectHome dynamically discovers the home directory.
func DetectHome() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "/data/data/com.termux/files/home"
}

// DetectJvmDir locates the JVM directory across Termux and standard Linux distributions.
func DetectJvmDir(prefix string) string {
	termuxJvm := filepath.Join(prefix, "lib", "jvm")
	if _, err := os.Stat(termuxJvm); err == nil {
		return termuxJvm
	}
	for _, cand := range []string{"/usr/lib/jvm", "/usr/local/lib/jvm"} {
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	return termuxJvm
}

// DetectPackageManager finds the native package manager available on the device.
func DetectPackageManager() string {
	for _, pm := range []string{"pkg", "apt-get", "apt", "pacman", "dnf", "zypper"} {
		if _, err := exec.LookPath(pm); err == nil {
			return pm
		}
	}
	return "pkg"
}

// DetectDeviceClass classifies the device type (Phone, Tablet, PC, Laptop, Chromebook, WSA, etc.).
func DetectDeviceClass(arch string, isTermux bool) string {
	if isTermux {
		switch arch {
		case "arm64", "aarch64":
			return "Android Phone / Tablet (ARM64 Termux)"
		case "arm", "armv7l", "armv8l":
			return "Android Device (32-bit ARM Termux)"
		case "amd64", "x86_64":
			return "PC / Laptop / Chromebook / WSA (x86_64 Termux)"
		case "386", "i686":
			return "x86 Android Emulator (32-bit Termux)"
		default:
			return "Termux (" + arch + ")"
		}
	}
	switch arch {
	case "amd64", "x86_64":
		return "PC / Laptop / Server (x86_64 Linux)"
	case "arm64", "aarch64":
		return "ARM64 SBC / PC / Server (Linux)"
	default:
		return "Linux (" + arch + ")"
	}
}

// GetPrimaryJavaHome returns the best active OpenJDK installation path.
func (cfg *Config) GetPrimaryJavaHome() string {
	prof := cfg.GetConfiguredJavaProfile()
	if prof.HostJavaDir != "" {
		hostPath := filepath.Join(cfg.JvmDir, prof.HostJavaDir)
		if _, err := os.Stat(filepath.Join(hostPath, "bin", "java")); err == nil {
			return hostPath
		}
	}

	candidates := []string{
		"java-21-openjdk",
		"java-25-openjdk",
		"java-17-openjdk",
		"default-java",
		"java-openjdk",
	}

	for _, cand := range candidates {
		p := filepath.Join(cfg.JvmDir, cand)
		if _, err := os.Stat(filepath.Join(p, "bin", "java")); err == nil {
			return p
		}
	}

	if jHome := os.Getenv("JAVA_HOME"); jHome != "" {
		return jHome
	}

	// Fallback to active java in PATH
	if p, err := exec.LookPath("java"); err == nil {
		if realP, err := filepath.EvalSymlinks(p); err == nil {
			return filepath.Dir(filepath.Dir(realP))
		}
	}

	return filepath.Join(cfg.JvmDir, "java-21-openjdk")
}

// DetectSystemMemoryMB reads /proc/meminfo to return total physical system RAM in megabytes.
func DetectSystemMemoryMB() int {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 4096 // Fallback default 4GB
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var kb int
				if _, err := fmt.Sscanf(fields[1], "%d", &kb); err == nil && kb > 0 {
					return kb / 1024
				}
			}
		}
	}
	return 4096
}

// DetectOptimalJvmArgs computes balanced heap sizing based on available physical memory.
// Prevents Android LMK (low memory killer) from terminating Gradle daemons during large dex/aapt tasks.
func DetectOptimalJvmArgs(memMB int) string {
	if memMB <= 0 {
		memMB = 4096
	}
	switch {
	case memMB < 3800:
		// Low-RAM devices (< 4GB): Lean heap to prevent OOM / LMK kills
		return "-Xmx1024m -XX:MaxMetaspaceSize=384m -XX:+UseG1GC"
	case memMB < 7500:
		// Mid-range devices (4GB - 6GB): Standard 2GB heap
		return "-Xmx2048m -XX:MaxMetaspaceSize=512m -XX:+UseG1GC"
	default:
		// High-end devices (8GB+ RAM): Expanded 3GB heap for ultra-fast builds
		return "-Xmx3072m -XX:MaxMetaspaceSize=768m -XX:+UseG1GC"
	}
}

// GetJavaOpts extracts the -Xmx heap flag from JvmArgs for shell export JAVA_OPTS.
func (cfg *Config) GetJavaOpts() string {
	for _, part := range strings.Fields(cfg.JvmArgs) {
		if strings.HasPrefix(part, "-Xmx") {
			return part
		}
	}
	return "-Xmx2048m"
}

// GetConfig initializes and returns the complete device environment configuration.
func GetConfig() (*Config, error) {
	prefix, isTermux := DetectPrefix()
	home := DetectHome()
	arch := runtime.GOARCH

	// Clean legacy cache path if present
	legacyCache := filepath.Join(home, ".android", ".andro_env_update_cache")
	newCache := filepath.Join(home, ".android", ".tadev_update_cache")
	if _, err := os.Stat(legacyCache); err == nil && !fileExists(newCache) {
		_ = os.Rename(legacyCache, newCache)
	}

	memMB := DetectSystemMemoryMB()
	optJvm := DetectOptimalJvmArgs(memMB)

	cfg := &Config{
		HomeDir:        home,
		AndroidDir:     filepath.Join(home, ".android"),
		SdkDir:         filepath.Join(home, ".android", "sdk"),
		CmdlineDir:     filepath.Join(home, ".android", "sdk", "cmdline-tools", "latest"),
		Prefix:         prefix,
		PrefixBin:      filepath.Join(prefix, "bin"),
		JvmDir:         DetectJvmDir(prefix),
		GradleDir:      filepath.Join(home, ".gradle"),
		GradleProps:    filepath.Join(home, ".gradle", "gradle.properties"),
		GradlewWrapper: filepath.Join(prefix, "bin", "gradlew"),
		UpdateCache:    newCache,
		ConfigDir:      filepath.Join(home, ".config", "tadev"),
		ConfigFile:     filepath.Join(home, ".config", "tadev", "config.yml"),
		BlockStart:     "### BEGIN TADEV (Managed by tadev) ###",
		BlockEnd:       "### END TADEV ###",
		Arch:           arch,
		DeviceClass:    DetectDeviceClass(arch, isTermux),
		IsTermux:       isTermux,
		PkgManager:     DetectPackageManager(),
		TotalMemMB:     memMB,
		OptimalJvmArgs: optJvm,
		JvmArgs:        optJvm,
		JavaProfile:    "21",
	}

	cfg.loadUserOverrides()

	return cfg, nil
}

func (cfg *Config) loadUserOverrides() {
	if !fileExists(cfg.ConfigFile) {
		return
	}

	file, err := os.Open(cfg.ConfigFile)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "sdk_root:") {
			val := strings.Trim(strings.TrimPrefix(line, "sdk_root:"), " \"'")
			if strings.HasPrefix(val, "~/") {
				val = filepath.Join(cfg.HomeDir, val[2:])
			}
			if val != "" {
				cfg.SdkDir = val
				cfg.CmdlineDir = filepath.Join(val, "cmdline-tools", "latest")
			}
		} else if strings.HasPrefix(line, "jvm_args:") {
			val := strings.Trim(strings.TrimPrefix(line, "jvm_args:"), " \"'")
			if val != "" {
				cfg.JvmArgs = val
			}
		} else if strings.HasPrefix(line, "profile:") || strings.HasPrefix(line, "java_profile:") {
			val := strings.Trim(strings.TrimPrefix(strings.TrimPrefix(line, "java_profile:"), "profile:"), " \"'")
			if val != "" {
				cfg.JavaProfile = NormalizeJavaProfileID(val)
			}
		}
	}
}

// GetConfiguredJavaProfile returns the active JavaProfile configured for the environment.
func (cfg *Config) GetConfiguredJavaProfile() JavaProfile {
	if p, ok := GetJavaProfile(cfg.JavaProfile); ok {
		return p
	}
	return DefaultJavaProfile()
}

// SaveJavaProfileSetting updates the configured Java profile inside ~/.config/tadev/config.yml.
func (cfg *Config) SaveJavaProfileSetting(profileID string) error {
	cfg.JavaProfile = NormalizeJavaProfileID(profileID)
	if err := os.MkdirAll(cfg.ConfigDir, 0755); err != nil {
		return err
	}

	var data []byte
	if fileExists(cfg.ConfigFile) {
		var err error
		data, err = os.ReadFile(cfg.ConfigFile)
		if err != nil {
			return err
		}
	} else {
		content := fmt.Sprintf(`version: "1.0.0"
android:
  sdk_root: "%s"
java:
  profile: "%s"
`, cfg.SdkDir, cfg.JavaProfile)
		return os.WriteFile(cfg.ConfigFile, []byte(content), 0644)
	}
	lines := strings.Split(string(data), "\n")
	found := false
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "profile:") {
			lines[i] = fmt.Sprintf("  profile: \"%s\"", cfg.JavaProfile)
			found = true
			break
		}
	}
	if !found {
		for i, l := range lines {
			if strings.TrimSpace(l) == "java:" {
				newLines := append(lines[:i+1], append([]string{fmt.Sprintf("  profile: \"%s\"", cfg.JavaProfile)}, lines[i+1:]...)...)
				lines = newLines
				found = true
				break
			}
		}
	}
	if !found {
		lines = append(lines, fmt.Sprintf("java:\n  profile: \"%s\"", cfg.JavaProfile))
	}
	return os.WriteFile(cfg.ConfigFile, []byte(strings.Join(lines, "\n")), 0644)
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
