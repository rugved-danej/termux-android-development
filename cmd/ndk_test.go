package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"tadev/pkg/env"
)

func TestNdkHelpers(t *testing.T) {
	tempHome := t.TempDir()
	sdkDir := filepath.Join(tempHome, "sdk")
	ndkDir := filepath.Join(sdkDir, "ndk", "27.2.12479018")
	_ = os.MkdirAll(ndkDir, 0755)

	cfg := &env.Config{
		SdkDir: sdkDir,
	}

	versions := env.ListInstalledNdkVersions(cfg)
	if len(versions) != 1 || versions[0] != "27.2.12479018" {
		t.Errorf("ListInstalledNdkVersions = %v, expected [27.2.12479018]", versions)
	}

	active := env.GetActiveNdkPath(cfg)
	if active != ndkDir {
		t.Errorf("GetActiveNdkPath = %s, want %s", active, ndkDir)
	}
}
