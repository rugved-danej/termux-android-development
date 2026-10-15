package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, got, tt.want)
		}
	}
}

func TestCalculateDirSize(t *testing.T) {
	tmp := t.TempDir()
	f1 := filepath.Join(tmp, "file1.txt")
	f2 := filepath.Join(tmp, "file2.txt")

	_ = os.WriteFile(f1, []byte("12345"), 0644) // 5 bytes
	_ = os.WriteFile(f2, []byte("12345"), 0644) // 5 bytes

	sz := calculateDirSize(tmp)
	if sz != 10 {
		t.Errorf("calculateDirSize() = %d, want 10", sz)
	}
}
