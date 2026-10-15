package env

import (
	"archive/zip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	GoogleRepoXML          = "https://dl.google.com/android/repository/repository2-3.xml"
	FallbackXML            = "https://dl.google.com/android/repository/repository2-1.xml"
	LatestJavaBuild        = "15641748" // Android SDK Command-line Tools 21.0 (pure Java)
	FallbackBuild          = "14742923" // Android SDK Command-line Tools 20.0 (pure Java)
	AltFallbackBuild       = "11076708" // Android SDK Command-line Tools 12.0 (pure Java LTS)
	MaxCompatibleJavaBuild = 15641748   // Builds above this contain x86_64 glibc ELF binaries incompatible with Android/Termux
)

type ProgressCallback func(downloaded, total int64)

// DownloadSdkFile downloads a remote file with progress reporting and automatic curl fallback.
func DownloadSdkFile(url, dest string, progress ProgressCallback) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		Proxy:           http.ProxyFromEnvironment,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   0, // Allow large streaming downloads
	}

	req, err := http.NewRequest("GET", url, nil)
	if err == nil {
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "*/*")

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			out, err := os.Create(dest)
			if err != nil {
				return err
			}
			defer out.Close()

			total := resp.ContentLength
			var downloaded int64
			buf := make([]byte, 64*1024)
			var lastProgress time.Time

			for {
				nr, er := resp.Body.Read(buf)
				if nr > 0 {
					nw, ew := out.Write(buf[:nr])
					if ew != nil {
						return ew
					}
					downloaded += int64(nw)
					now := time.Now()
					if progress != nil && (now.Sub(lastProgress) >= 100*time.Millisecond || downloaded >= total) {
						lastProgress = now
						progress(downloaded, total)
					}
				}
				if er != nil {
					if er != io.EOF {
						return er
					}
					break
				}
			}
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// High-reliability fallback to native curl (handles Termux SSL certificates and proxies)
	if _, err := exec.LookPath("curl"); err == nil {
		cmd := exec.Command("curl", "-fL", "--retry", "3", "--connect-timeout", "20", "-A", "Mozilla/5.0", "-o", dest, url)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			return fmt.Errorf("curl download failed: %v (%s)", err, string(out))
		}
	}

	// Fallback to wget if curl is not installed
	if _, err := exec.LookPath("wget"); err == nil {
		cmd := exec.Command("wget", "-q", "-c", "--tries=3", "--timeout=20", "-U", "Mozilla/5.0", "-O", dest, url)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to download from %s: %v", url, err)
}

func FetchLatestCmdlineVersion() string {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", GoogleRepoXML, nil)
	if err != nil {
		return LatestJavaBuild
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	res, err := client.Do(req)
	if err != nil || res.StatusCode != 200 {
		if res != nil {
			res.Body.Close()
		}
		req2, err := http.NewRequest("GET", FallbackXML, nil)
		if err != nil {
			return LatestJavaBuild
		}
		req2.Header.Set("User-Agent", "Mozilla/5.0")
		res, err = client.Do(req2)
		if err != nil || res.StatusCode != 200 {
			if res != nil {
				res.Body.Close()
			}
			return LatestJavaBuild
		}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return LatestJavaBuild
	}

	re := regexp.MustCompile(`commandlinetools-linux-(\d+)_latest\.zip`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if len(matches) == 0 {
		return LatestJavaBuild
	}

	var versions []int
	for _, m := range matches {
		if len(m) > 1 {
			if v, err := strconv.Atoi(m[1]); err == nil {
				// Only select builds known to be compatible pure Java without x86_64 ELF glibc wrappers
				if v <= MaxCompatibleJavaBuild {
					versions = append(versions, v)
				}
			}
		}
	}
	if len(versions) == 0 {
		return LatestJavaBuild
	}
	sort.Ints(versions)
	return strconv.Itoa(versions[len(versions)-1])
}

func SetupCmdlineTools(cfg *Config, force bool, progress ProgressCallback) (string, error) {
	latestVer := FetchLatestCmdlineVersion()
	if latestVer == "" {
		latestVer = LatestJavaBuild
	}

	versionFile := filepath.Join(cfg.SdkDir, "cmdline-tools", ".build_version")
	currentVer := ""
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVer = string(data)
	}

	binDir := filepath.Join(cfg.CmdlineDir, "bin")
	if !force && currentVer == latestVer && dirExists(binDir) {
		return latestVer, nil
	}

	cmdlineToolsDir := filepath.Join(cfg.SdkDir, "cmdline-tools")
	if err := os.MkdirAll(cmdlineToolsDir, 0755); err != nil {
		return "", err
	}

	zipPath := filepath.Join(cmdlineToolsDir, "cmdline-tools.zip")
	_ = os.Remove(zipPath)

	// Try download candidates in prioritized order
	candidates := []string{latestVer}
	for _, v := range []string{LatestJavaBuild, FallbackBuild, AltFallbackBuild} {
		found := false
		for _, c := range candidates {
			if c == v {
				found = true
				break
			}
		}
		if !found {
			candidates = append(candidates, v)
		}
	}

	var downloadErr error
	var successfulVer string

	for _, ver := range candidates {
		url := fmt.Sprintf("https://dl.google.com/android/repository/commandlinetools-linux-%s_latest.zip", ver)
		if err := DownloadSdkFile(url, zipPath, progress); err == nil {
			successfulVer = ver
			downloadErr = nil
			break
		} else {
			downloadErr = err
		}
	}

	if downloadErr != nil {
		return "", fmt.Errorf("failed to download commandlinetools: %w", downloadErr)
	}

	latestVer = successfulVer

	// Extract zip
	if err := unzipFile(zipPath, cmdlineToolsDir); err != nil {
		return "", err
	}
	_ = os.Remove(zipPath)

	extractedDir := filepath.Join(cmdlineToolsDir, "cmdline-tools")
	if dirExists(extractedDir) {
		_ = os.RemoveAll(cfg.CmdlineDir)
		if err := os.Rename(extractedDir, cfg.CmdlineDir); err != nil {
			return "", err
		}
	}

	// Remove any incompatible x86_64 binary if present so sdkmanager runs in pure Java mode
	_ = os.Remove(filepath.Join(cfg.CmdlineDir, "bin", "android"))

	_ = os.WriteFile(versionFile, []byte(latestVer), 0644)

	// Fix shebangs recursively
	FixAllShebangs(cfg)

	return latestVer, nil
}

func isScript(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	header := make([]byte, 2)
	n, err := f.Read(header)
	return err == nil && n == 2 && header[0] == '#' && header[1] == '!'
}

// FixAllShebangs recursively updates shebangs for all tools installed under SDK.
func FixAllShebangs(cfg *Config) {
	hasTermuxFix := false
	if _, err := exec.LookPath("termux-fix-shebang"); err == nil {
		hasTermuxFix = true
	}

	dirs := []string{
		filepath.Join(cfg.CmdlineDir, "bin"),
		filepath.Join(cfg.SdkDir, "platform-tools"),
	}

	btBase := filepath.Join(cfg.SdkDir, "build-tools")
	if entries, err := os.ReadDir(btBase); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				dirs = append(dirs, filepath.Join(btBase, e.Name()))
			}
		}
	}

	for _, d := range dirs {
		if files, err := os.ReadDir(d); err == nil {
			for _, f := range files {
				if !f.IsDir() {
					p := filepath.Join(d, f.Name())
					if hasTermuxFix && isScript(p) {
						_ = exec.Command("termux-fix-shebang", p).Run()
					}
					_ = os.Chmod(p, 0755)
				}
			}
		}
	}

	if activeNdk := GetActiveNdkPath(cfg); activeNdk != "" {
		FixNdkShebangs(cfg, activeNdk)
	}
}

// InstallSdkPackages accepts licenses and installs designated SDK platforms/tools.
func InstallSdkPackages(cfg *Config, pkgs ...string) error {
	sdkmanager := filepath.Join(cfg.CmdlineDir, "bin", "sdkmanager")
	if !fileExists(sdkmanager) {
		return fmt.Errorf("sdkmanager not found at %s. Please run 'tadev setup' first", sdkmanager)
	}

	if len(pkgs) == 0 {
		pkgs = []string{"platform-tools", "platforms;android-36", "platforms;android-35", "build-tools;36.0.0"}
	}

	javaHome := cfg.GetPrimaryJavaHome()
	env := os.Environ()
	env = append(env,
		"JAVA_HOME="+javaHome,
		"ANDROID_HOME="+cfg.SdkDir,
		"ANDROID_SDK_ROOT="+cfg.SdkDir,
		"PATH="+filepath.Join(cfg.CmdlineDir, "bin")+":"+filepath.Join(javaHome, "bin")+":"+os.Getenv("PATH"),
	)

	// Automatically accept licenses: yes | sdkmanager --sdk_root=... --licenses
	cmdLicenses := exec.Command("sh", "-c", fmt.Sprintf("yes | %s --sdk_root=%s --licenses", sdkmanager, cfg.SdkDir))
	cmdLicenses.Env = env
	_ = cmdLicenses.Run()

	args := append([]string{"--sdk_root=" + cfg.SdkDir}, pkgs...)
	cmdInstall := exec.Command(sdkmanager, args...)
	cmdInstall.Env = env
	if out, err := cmdInstall.CombinedOutput(); err != nil {
		return fmt.Errorf("sdkmanager failed: %v\n%s", err, string(out))
	}

	FixAllShebangs(cfg)
	return nil
}

// ListInstalledSdkPackages returns installed platforms and build tools.
func ListInstalledSdkPackages(cfg *Config) (platforms []string, buildTools []string) {
	platsDir := filepath.Join(cfg.SdkDir, "platforms")
	if entries, err := os.ReadDir(platsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				platforms = append(platforms, e.Name())
			}
		}
	}

	btDir := filepath.Join(cfg.SdkDir, "build-tools")
	if entries, err := os.ReadDir(btDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				buildTools = append(buildTools, e.Name())
			}
		}
	}
	return
}

// ListInstalledNdkVersions returns sorted list of installed NDK versions.
func ListInstalledNdkVersions(cfg *Config) []string {
	var versions []string
	ndkBase := filepath.Join(cfg.SdkDir, "ndk")
	if entries, err := os.ReadDir(ndkBase); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				versions = append(versions, e.Name())
			}
		}
	}
	bundlePath := filepath.Join(cfg.SdkDir, "ndk-bundle")
	if info, err := os.Stat(bundlePath); err == nil && info.IsDir() {
		versions = append(versions, "ndk-bundle")
	}
	sort.Strings(versions)
	return versions
}

// GetActiveNdkPath returns the latest installed NDK path, or empty string if none.
func GetActiveNdkPath(cfg *Config) string {
	versions := ListInstalledNdkVersions(cfg)
	if len(versions) == 0 {
		return ""
	}
	latest := versions[len(versions)-1]
	if latest == "ndk-bundle" {
		return filepath.Join(cfg.SdkDir, "ndk-bundle")
	}
	return filepath.Join(cfg.SdkDir, "ndk", latest)
}

// FixNdkShebangs walks the NDK directory and updates shebangs on all scripts and tools.
func FixNdkShebangs(cfg *Config, ndkPath string) {
	if !dirExists(ndkPath) {
		return
	}
	hasTermuxFix := false
	if _, err := exec.LookPath("termux-fix-shebang"); err == nil {
		hasTermuxFix = true
	}

	_ = filepath.Walk(ndkPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		name := info.Name()
		if isScript(path) || strings.HasSuffix(name, ".sh") || strings.HasPrefix(name, "ndk-") {
			if hasTermuxFix && isScript(path) {
				_ = exec.Command("termux-fix-shebang", path).Run()
			}
			_ = os.Chmod(path, 0755)
		}
		return nil
	})
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func unzipFile(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	destClean := filepath.Clean(dest) + string(os.PathSeparator)

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		fpathClean := filepath.Clean(fpath)
		if !strings.HasPrefix(fpathClean, destClean) && fpathClean != filepath.Clean(dest) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
