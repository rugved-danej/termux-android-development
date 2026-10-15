package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"tadev/pkg/env"
	"tadev/pkg/ui"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getAuthToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	if t := os.Getenv("GH_TOKEN"); t != "" {
		return t
	}
	if out, err := exec.Command("gh", "auth", "token").Output(); err == nil {
		tok := strings.TrimSpace(string(out))
		if tok != "" {
			return tok
		}
	}
	return ""
}

// RunSelfUpdate checks for and applies upgrades to the tadev CLI binary itself.
func RunSelfUpdate(currentVersion string, force bool) error {
	ui.PrintBanner()

	cfg, err := env.GetConfig()
	if err != nil {
		return err
	}

	sp := ui.NewSpinner("Checking for latest tadev CLI release on GitHub...")
	sp.Start()

	token := getAuthToken()
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/rugved-danej/termux-android-development/releases/latest", nil)
	req.Header.Set("User-Agent", "tadev-updater")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	var rel githubRelease
	if err == nil && resp.StatusCode == 200 {
		_ = json.NewDecoder(resp.Body).Decode(&rel)
		resp.Body.Close()
	} else {
		if resp != nil {
			resp.Body.Close()
		}
		// Try gh release view
		if out, err := exec.Command("gh", "release", "view", "--repo", "rugved-danej/termux-android-development", "--json", "tagName").Output(); err == nil {
			var ghOut struct {
				TagName string `json:"tagName"`
			}
			if json.Unmarshal(out, &ghOut) == nil && ghOut.TagName != "" {
				rel.TagName = ghOut.TagName
			}
		}
	}

	latestTag := strings.TrimPrefix(rel.TagName, "v")
	if latestTag == "" {
		latestTag = "1.0.0"
	}

	currentNorm := strings.TrimPrefix(currentVersion, "v")
	if !force && currentNorm == latestTag {
		sp.StopSuccess(fmt.Sprintf("tadev is already on the latest version (%s).", currentVersion))
		return nil
	}

	sp.StopSuccess(fmt.Sprintf("Release available: %s -> %s", currentVersion, latestTag))

	targetName := fmt.Sprintf("tadev-android-%s", cfg.Arch)
	destPath := filepath.Join(cfg.PrefixBin, "tadev")
	tempPath := filepath.Join(cfg.PrefixBin, "tadev.tmp")

	spDownload := ui.NewSpinner(fmt.Sprintf("Downloading %s...", targetName))
	spDownload.Start()

	// Try downloading via gh CLI first if authenticated
	downloadedViaGh := false
	if _, err := exec.LookPath("gh"); err == nil {
		tagArg := "v" + latestTag
		cmdGh := exec.Command("gh", "release", "download", tagArg, "-p", targetName, "-O", tempPath, "--repo", "rugved-danej/termux-android-development")
		if err := cmdGh.Run(); err == nil {
			downloadedViaGh = true
		}
	}

	if !downloadedViaGh {
		downloadUrl := fmt.Sprintf("https://github.com/rugved-danej/termux-android-development/releases/download/v%s/%s", latestTag, targetName)
		reqDown, _ := http.NewRequest("GET", downloadUrl, nil)
		if token != "" {
			reqDown.Header.Set("Authorization", "Bearer "+token)
		}
		respDown, err := client.Do(reqDown)
		if err != nil || respDown.StatusCode != 200 {
			if respDown != nil {
				respDown.Body.Close()
			}
			downloadUrl = fmt.Sprintf("https://github.com/rugved-danej/termux-android-development/releases/latest/download/%s", targetName)
			reqDown, _ = http.NewRequest("GET", downloadUrl, nil)
			if token != "" {
				reqDown.Header.Set("Authorization", "Bearer "+token)
			}
			respDown, err = client.Do(reqDown)
			if err != nil || respDown.StatusCode != 200 {
				spDownload.StopFail("Failed to download release binary from GitHub")
				return fmt.Errorf("download error from %s", downloadUrl)
			}
		}
		defer respDown.Body.Close()

		out, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			spDownload.StopFail(err.Error())
			return err
		}

		total := respDown.ContentLength
		var downloaded int64
		buf := make([]byte, 32*1024)
		var lastProgress time.Time
		for {
			nr, er := respDown.Body.Read(buf)
			if nr > 0 {
				nw, ew := out.Write(buf[0:nr])
				if ew != nil {
					out.Close()
					_ = os.Remove(tempPath)
					spDownload.StopFail(ew.Error())
					return ew
				}
				downloaded += int64(nw)
				now := time.Now()
				if total > 0 && (now.Sub(lastProgress) >= 100*time.Millisecond || downloaded >= total) {
					lastProgress = now
					bar := ui.RenderProgressBar(downloaded, total, 10)
					mbDown := float64(downloaded) / (1024 * 1024)
					mbTot := float64(total) / (1024 * 1024)
					spDownload.UpdateMessage(fmt.Sprintf("Downloading %s: %.1fMB/%.1fMB %s", targetName, mbDown, mbTot, bar))
				}
			}
			if er != nil {
				if er != io.EOF {
					out.Close()
					_ = os.Remove(tempPath)
					spDownload.StopFail(er.Error())
					return er
				}
				break
			}
		}
		out.Close()
	}

	// Replace active binary atomically
	_ = os.Remove(destPath)
	if err := os.Rename(tempPath, destPath); err != nil {
		spDownload.StopFail("Failed to replace binary: " + err.Error())
		return err
	}
	_ = os.Chmod(destPath, 0755)

	// Recreate relative symlink
	altPath := filepath.Join(cfg.PrefixBin, "termux-android-dev")
	_ = os.Remove(altPath)
	_ = os.Symlink("tadev", altPath)

	spDownload.StopSuccess(fmt.Sprintf("tadev updated successfully to version %s!", latestTag))
	return nil
}
