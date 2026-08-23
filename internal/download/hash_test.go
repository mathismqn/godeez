package download

import (
	"path/filepath"
	"testing"
)

func TestHashFile(t *testing.T) {
	path := writeFile(t, filepath.Join(t.TempDir(), "file.txt"), "hello world")

	got, size, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile: %v", err)
	}

	if got != helloWorldHash {
		t.Errorf("hashFile() hash = %s, want %s", got, helloWorldHash)
	}
	if size != int64(len("hello world")) {
		t.Errorf("hashFile() size = %d, want %d", size, len("hello world"))
	}
}

func TestHashFileMissing(t *testing.T) {
	if _, _, err := hashFile(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("expected error for missing file")
	}
}
