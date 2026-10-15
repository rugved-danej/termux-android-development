package env

import (
	"strings"
	"testing"
)

func TestDetectOptimalJvmArgs(t *testing.T) {
	tests := []struct {
		name       string
		memMB      int
		wantSubstr string
	}{
		{"Low RAM device (< 4GB)", 2048, "-Xmx1024m"},
		{"Budget 3GB device", 3072, "-Xmx1024m"},
		{"Standard 4GB device", 4096, "-Xmx2048m"},
		{"Mid-range 6GB device", 6144, "-Xmx2048m"},
		{"High-end 8GB device", 8192, "-Xmx3072m"},
		{"Flagship 12GB device", 12288, "-Xmx3072m"},
		{"Zero fallback", 0, "-Xmx2048m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectOptimalJvmArgs(tt.memMB)
			if !strings.Contains(got, tt.wantSubstr) {
				t.Errorf("DetectOptimalJvmArgs(%d) = %s, expected to contain %s", tt.memMB, got, tt.wantSubstr)
			}
		})
	}
}

func TestGetJavaOpts(t *testing.T) {
	cfg := &Config{
		JvmArgs: "-Xmx3072m -XX:MaxMetaspaceSize=768m -XX:+UseG1GC",
	}
	opts := cfg.GetJavaOpts()
	if opts != "-Xmx3072m" {
		t.Errorf("GetJavaOpts() = %s, want -Xmx3072m", opts)
	}

	cfgEmpty := &Config{JvmArgs: ""}
	optsDefault := cfgEmpty.GetJavaOpts()
	if optsDefault != "-Xmx2048m" {
		t.Errorf("GetJavaOpts() on empty = %s, want -Xmx2048m", optsDefault)
	}
}

func TestDetectSystemMemoryMB(t *testing.T) {
	mem := DetectSystemMemoryMB()
	if mem <= 0 {
		t.Errorf("DetectSystemMemoryMB() returned invalid %d", mem)
	}
}
