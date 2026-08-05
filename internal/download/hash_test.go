package download

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHashFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile: %v", err)
	}

	want := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if got != want {
		t.Errorf("hashFile() = %s, want %s", got, want)
	}
}

func TestHashFileMissing(t *testing.T) {
	if _, err := hashFile(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestHashIndexFind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "track.mp3")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	index, err := newHashIndex(context.Background(), dir)
	if err != nil {
		t.Fatalf("newHashIndex: %v", err)
	}

	hash := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	found, ok := index.find(hash)
	if !ok || found != path {
		t.Errorf("find(%s) = %q, %v; want %q, true", hash, found, ok, path)
	}

	if _, ok := index.find("deadbeef"); ok {
		t.Error("find() reported a match for an unknown hash")
	}
}

func TestHashIndexCanceledContext(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := newHashIndex(ctx, dir); err == nil {
		t.Error("expected error for canceled context")
	}
}
