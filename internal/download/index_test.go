package download

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileIndexFind(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "sub", "track.mp3"), "hello world")

	index, err := newFileIndex(context.Background(), dir)
	if err != nil {
		t.Fatalf("newFileIndex: %v", err)
	}

	size := int64(len("hello world"))
	found, foundSize, ok := index.find(helloWorldHash, size)
	if !ok || found != path || foundSize != size {
		t.Errorf("find(hash, %d) = %q, %d, %v; want %q, %d, true", size, found, foundSize, ok, path, size)
	}

	if _, _, ok := index.find("deadbeef", size); ok {
		t.Error("find() reported a match for an unknown hash")
	}
	if _, _, ok := index.find("", 0); ok {
		t.Error("find() reported a match for an empty hash")
	}
}

// TestFileIndexFindWithoutSize covers a record written before the ledger
// tracked sizes: every file is a candidate, and the size found is reported
// back for the caller to record.
func TestFileIndexFindWithoutSize(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "track.flac"), "hello world")
	writeFile(t, filepath.Join(dir, "other.mp3"), "something else entirely")

	index, err := newFileIndex(context.Background(), dir)
	if err != nil {
		t.Fatalf("newFileIndex: %v", err)
	}

	found, foundSize, ok := index.find(helloWorldHash, 0)
	if !ok || found != path || foundSize != int64(len("hello world")) {
		t.Errorf("find(hash, 0) = %q, %d, %v; want %q, %d, true", found, foundSize, ok, path, len("hello world"))
	}
}

// TestFileIndexSizeNarrowsReads is the point of the whole index: a lookup with
// a known size must not read the files that could not possibly match.
func TestFileIndexSizeNarrowsReads(t *testing.T) {
	dir := t.TempDir()
	want := writeFile(t, filepath.Join(dir, "track.mp3"), "hello world")
	for _, name := range []string{"a.mp3", "b.flac", "c.wav"} {
		writeFile(t, filepath.Join(dir, name), "a file of an entirely different length")
	}

	index, err := newFileIndex(context.Background(), dir)
	if err != nil {
		t.Fatalf("newFileIndex: %v", err)
	}
	if len(index.bySize) != 2 {
		t.Fatalf("bySize has %d sizes, want 2", len(index.bySize))
	}

	found, _, ok := index.find(helloWorldHash, int64(len("hello world")))
	if !ok || found != want {
		t.Fatalf("find() = %q, %v; want %q, true", found, ok, want)
	}

	if len(index.hashes) != 1 {
		t.Errorf("hashed %d files, want 1: the size filter should have rejected the rest without reading them", len(index.hashes))
	}
}

// TestFileIndexSkipsNonCandidates keeps the walk off everything that cannot
// be a recorded download, most importantly the ledger, which bbolt holds open
// and mutating while this runs.
func TestFileIndexSkipsNonCandidates(t *testing.T) {
	dir := t.TempDir()
	kept := writeFile(t, filepath.Join(dir, "album", "track.mp3"), "hello world")

	writeFile(t, filepath.Join(dir, ".tracks.db"), "hello world")
	writeFile(t, filepath.Join(dir, ".godeez-123.part"), "hello world")
	writeFile(t, filepath.Join(dir, "album", "cover.jpg"), "hello world")
	writeFile(t, filepath.Join(dir, ".hidden", "track.mp3"), "hello world")

	index, err := newFileIndex(context.Background(), dir)
	if err != nil {
		t.Fatalf("newFileIndex: %v", err)
	}

	var indexed []string
	for _, paths := range index.bySize {
		indexed = append(indexed, paths...)
	}

	if len(indexed) != 1 || indexed[0] != kept {
		t.Errorf("indexed %v, want only %q", indexed, kept)
	}
}

func TestFileIndexCanceledContext(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.mp3"), "x")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := newFileIndex(ctx, dir); err == nil {
		t.Error("expected error for canceled context")
	}
}

// TestFileIndexUnreadableFile makes sure a file that cannot be read is simply
// never a match, rather than breaking the lookup for everything else.
func TestFileIndexUnreadableFile(t *testing.T) {
	dir := t.TempDir()
	bad := writeFile(t, filepath.Join(dir, "bad.mp3"), "hello world")
	good := writeFile(t, filepath.Join(dir, "good.mp3"), "hello world")

	if err := os.Chmod(bad, 0000); err != nil {
		t.Skipf("cannot make a file unreadable here: %v", err)
	}
	t.Cleanup(func() { os.Chmod(bad, 0644) })

	// Running as root defeats the permission bits entirely, which would make
	// this assert the opposite of what it means to.
	if f, err := os.Open(bad); err == nil {
		f.Close()
		t.Skip("files are readable regardless of mode here")
	}

	index, err := newFileIndex(context.Background(), dir)
	if err != nil {
		t.Fatalf("newFileIndex: %v", err)
	}

	found, _, ok := index.find(helloWorldHash, int64(len("hello world")))
	if !ok || found != good {
		t.Errorf("find() = %q, %v; want %q, true", found, ok, good)
	}
}
