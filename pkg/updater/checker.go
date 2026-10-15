package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

// FetchLatestCliVersion checks GitHub releases for the latest published tadev version.
func FetchLatestCliVersion() string {
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/rugved-danej/termux-android-development/releases/latest", nil)
	req.Header.Set("User-Agent", "tadev-updater")
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	} else if tok := os.Getenv("GH_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	} else if out, err := exec.Command("gh", "auth", "token").Output(); err == nil {
		tok := strings.TrimSpace(string(out))
		if tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
	}

	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		var rel struct {
			TagName string `json:"tag_name"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&rel)
		resp.Body.Close()
		return strings.TrimPrefix(rel.TagName, "v")
	}
	if resp != nil {
		resp.Body.Close()
	}

	if out, err := exec.Command("gh", "release", "view", "--repo", "rugved-danej/termux-android-development", "--json", "tagName").Output(); err == nil {
		var ghOut struct {
			TagName string `json:"tagName"`
		}
		if json.Unmarshal(out, &ghOut) == nil && ghOut.TagName != "" {
			return strings.TrimPrefix(ghOut.TagName, "v")
		}
	}
	return ""
}

func isNewer(current, latest string) bool {
	cParts := strings.Split(strings.TrimPrefix(current, "v"), ".")
	lParts := strings.Split(strings.TrimPrefix(latest, "v"), ".")
	for i := 0; i < len(cParts) && i < len(lParts); i++ {
		cV, _ := strconv.Atoi(cParts[i])
		lV, _ := strconv.Atoi(lParts[i])
		if lV > cV {
			return true
		} else if lV < cV {
			return false
		}
	}
	return len(lParts) > len(cParts)
}

// CheckUpdates performs a diagnostic check across Google repos, package repositories, and GitHub releases.
func CheckUpdates(cfg *env.Config, currentVersion string, quiet bool) error {
	if quiet {
		if data, err := os.ReadFile(cfg.UpdateCache); err == nil {
			if ts, err := strconv.ParseInt(string(data), 10, 64); err == nil {
				// Only check once every 7 days (604800 seconds) in quiet background mode
				if time.Now().Unix()-ts < 604800 {
					return nil
				}
			}
		}
	}

	var sp *ui.Spinner
	if !quiet {
		sp = ui.NewSpinner("Checking for updates from Google repository, Termux repositories, and GitHub...")
		sp.Start()
	}

	var updates []string

	// 1. Check tadev CLI release
	latestCli := FetchLatestCliVersion()
	currentNorm := strings.TrimPrefix(currentVersion, "v")
	if latestCli != "" && currentNorm != "" && isNewer(currentNorm, latestCli) {
		updates = append(updates, fmt.Sprintf("tadev CLI Tool: %s -> %s (Run 'tadev self-update' or 'tadev update')", currentVersion, latestCli))
	}

	// 2. Check Android Command-Line tools build
	latestCmdline := env.FetchLatestCmdlineVersion()
	versionFile := filepath.Join(cfg.SdkDir, "cmdline-tools", ".build_version")
	currentCmdline := ""
	if data, err := os.ReadFile(versionFile); err == nil {
		currentCmdline = string(data)
	}
	if latestCmdline != "" && currentCmdline != "" && latestCmdline != currentCmdline {
		updates = append(updates, fmt.Sprintf("Android Command-Line Tools: build %s -> build %s", currentCmdline, latestCmdline))
	}

	// 3. Check native packages
	pkgUpdates, _ := env.CheckPackageUpdates()
	for _, p := range pkgUpdates {
		updates = append(updates, "Package update: "+p)
	}

	_ = os.WriteFile(cfg.UpdateCache, []byte(strconv.FormatInt(time.Now().Unix(), 10)), 0644)

	if !quiet {
		if len(updates) > 0 {
			sp.StopSuccess(fmt.Sprintf("%d update(s) available!", len(updates)))
		} else {
			sp.StopSuccess("Everything is up to date!")
		}
	}

	if len(updates) > 0 {
		if quiet {
			fmt.Println("\n" + ui.Yellow + ui.Bold + "[*] [tadev] Updates available for Android SDK / tadev CLI!" + ui.Reset)
			fmt.Println("   Run " + ui.Cyan + "'tadev check-updates'" + ui.Reset + " or " + ui.Cyan + "'tadev update'" + ui.Reset + " to view and apply.")
		} else {
			fmt.Println()
			ui.PrintBox("AVAILABLE UPDATES", updates)
			fmt.Printf("\nRun %stadev update%s to apply all updates.\n\n", ui.Green+ui.Bold, ui.Reset)
		}
	} else if !quiet {
		fmt.Printf("\n%sYour Android SDK, OpenJDK, build tools, and tadev CLI are all on the latest versions.%s\n\n", ui.Green, ui.Reset)
	}

	return nil
}

// ApplyUpdates upgrades the packages, Command-Line Tools, and Android SDK platforms/build-tools.
func ApplyUpdates(cfg *env.Config) error {
	ui.PrintBanner()

	prof := cfg.GetConfiguredJavaProfile()
	sp1 := ui.NewSpinner(fmt.Sprintf("Updating core development packages (%s, Gradle, AAPT2)...", prof.ID))
	sp1.Start()
	if err := env.InstallTermuxPackagesForProfile(prof); err != nil {
		sp1.StopFail("Failed to update packages: " + err.Error())
	} else {
		sp1.StopSuccess("Packages upgraded successfully.")
	}

	sp2 := ui.NewSpinner("Checking & updating Android Command-Line Tools...")
	sp2.Start()
	ver, err := env.SetupCmdlineTools(cfg, false, nil)
	if err != nil {
		sp2.StopFail("Failed to update Command-Line tools: " + err.Error())
	} else {
		sp2.StopSuccess("Command-Line Tools build " + ver + " verified.")
	}

	sp3 := ui.NewSpinner("Updating Android SDK platforms & build-tools via sdkmanager...")
	sp3.Start()
	sdkmanager := filepath.Join(cfg.CmdlineDir, "bin", "sdkmanager")
	if _, err := os.Stat(sdkmanager); err == nil {
		cmd := exec.Command(sdkmanager, "--update")
		cmd.Env = append(os.Environ(),
			"JAVA_HOME="+cfg.GetPrimaryJavaHome(),
			"ANDROID_HOME="+cfg.SdkDir,
			"PATH="+filepath.Join(cfg.CmdlineDir, "bin")+":"+os.Getenv("PATH"),
		)
		_ = cmd.Run()
		sp3.StopSuccess("Android SDK platforms and build-tools are up-to-date.")
	} else {
		sp3.StopFail("sdkmanager not found at " + sdkmanager)
	}

	env.FixAllShebangs(cfg)
	_ = env.ConfigureGradleProperties(cfg)
	_ = env.InstallGradlewWrapper(cfg)

	_ = os.WriteFile(cfg.UpdateCache, []byte(strconv.FormatInt(time.Now().Unix(), 10)), 0644)

	fmt.Printf("\n%s[OK] All Android development environment updates applied successfully!%s\n\n", ui.Green+ui.Bold, ui.Reset)
	return nil
}
