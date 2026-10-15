package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ConfigureGradleProperties(cfg *Config) error {
	prof := cfg.GetConfiguredJavaProfile()
	var paths []string
	for _, d := range prof.ToolchainDirs {
		paths = append(paths, filepath.Join(cfg.JvmDir, d))
	}
	return ConfigureGradlePropertiesWithPaths(cfg, paths)
}

func ConfigureGradlePropertiesWithPaths(cfg *Config, toolchainPaths []string) error {
	if err := os.MkdirAll(cfg.GradleDir, 0755); err != nil {
		return err
	}

	content := ""
	if data, err := os.ReadFile(cfg.GradleProps); err == nil {
		content = string(data)
	}

	content = RemoveManagedBlock(content, cfg.BlockStart, cfg.BlockEnd)

	if len(toolchainPaths) == 0 {
		toolchainPaths = []string{
			filepath.Join(cfg.JvmDir, "java-25-openjdk"),
			filepath.Join(cfg.JvmDir, "java-21-openjdk"),
		}
	}

	pathsJoined := strings.Join(toolchainPaths, ",")

	jvmArgs := cfg.JvmArgs
	if jvmArgs == "" {
		jvmArgs = "-Xmx2048m -XX:MaxMetaspaceSize=512m -XX:+UseG1GC"
	}

	block := fmt.Sprintf(`%s
# Native ARM64 AAPT2 Override for Termux
android.aapt2FromMavenOverride=%s/aapt2
# Local JVM Toolchain search paths for multi-version JDK projects
org.gradle.java.installations.paths=%s
org.gradle.jvmargs=%s
%s
`, cfg.BlockStart, cfg.PrefixBin, pathsJoined, jvmArgs, cfg.BlockEnd)

	newContent := strings.TrimSpace(content)
	if newContent != "" {
		newContent += "\n\n"
	}
	newContent += block

	return os.WriteFile(cfg.GradleProps, []byte(newContent), 0644)
}

func CleanGradleProperties(cfg *Config) error {
	if _, err := os.Stat(cfg.GradleProps); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(cfg.GradleProps)
	if err != nil {
		return err
	}

	cleaned := RemoveManagedBlock(string(data), cfg.BlockStart, cfg.BlockEnd)
	var finalLines []string
	for _, l := range strings.Split(cleaned, "\n") {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "android.aapt2FromMavenOverride") ||
			strings.HasPrefix(t, "org.gradle.java.installations.paths") {
			continue
		}
		if t != "" {
			finalLines = append(finalLines, l)
		}
	}

	if len(finalLines) == 0 {
		return os.Remove(cfg.GradleProps)
	}
	return os.WriteFile(cfg.GradleProps, []byte(strings.Join(finalLines, "\n")+"\n"), 0644)
}

func InstallGradlewWrapper(cfg *Config) error {
	script := `#!/bin/sh
# Smart gradlew runner for Termux & Linux (bypasses noexec mount on Android storage)
d="$PWD"
while [ "$d" != "/" ] && [ "$d" != "" ]; do
    if [ -f "$d/gradlew" ]; then
        if command -v bash >/dev/null 2>&1; then
            exec bash "$d/gradlew" "$@"
        else
            exec sh "$d/gradlew" "$@"
        fi
    fi
    d="$(dirname "$d")"
done

exec gradle "$@"
`
	return os.WriteFile(cfg.GradlewWrapper, []byte(script), 0755)
}

func RemoveGradlewWrapper(cfg *Config) error {
	if _, err := os.Stat(cfg.GradlewWrapper); err == nil {
		return os.Remove(cfg.GradlewWrapper)
	}
	return nil
}

func RemoveManagedBlock(content, start, end string) string {
	for {
		sIdx := strings.Index(content, start)
		if sIdx == -1 {
			break
		}
		eIdx := strings.Index(content[sIdx:], end)
		if eIdx == -1 {
			break
		}
		eIdx += sIdx + len(end)
		// Also clean trailing newline
		if eIdx < len(content) && content[eIdx] == '\n' {
			eIdx++
		}
		content = content[:sIdx] + content[eIdx:]
	}
	return content
}
