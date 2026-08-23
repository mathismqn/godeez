package download

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mathismqn/godeez/internal/store"
)

func TestShouldSkipDownloadExistingPath(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "track.mp3"), "hello world")

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Quality: "MP3_320", Path: path, Hash: helloWorldHash, Size: 11}); err != nil {
		t.Fatal(err)
	}

	got, ok := d.shouldSkipDownload(context.Background(), "1", "MP3_320")
	if !ok || got != path {
		t.Errorf("shouldSkipDownload() = %q, %v; want %q, true", got, ok, path)
	}
	if d.fileIndex != nil {
		t.Error("scanned the library for a file that was where the ledger said")
	}
}

// TestShouldSkipDownloadRetaggedFile guards the decision not to check the size
// of a file that is present: another tagger rewriting the bytes must not cost
// the user a re-download.
func TestShouldSkipDownloadRetaggedFile(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "track.mp3"), "these bytes have since been rewritten")

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Quality: "MP3_320", Path: path, Hash: helloWorldHash, Size: 11}); err != nil {
		t.Fatal(err)
	}

	if got, ok := d.shouldSkipDownload(context.Background(), "1", "MP3_320"); !ok || got != path {
		t.Errorf("shouldSkipDownload() = %q, %v; want %q, true", got, ok, path)
	}
}

func TestShouldSkipDownloadDifferentQuality(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "track.mp3"), "hello world")

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Quality: "MP3_320", Path: path, Hash: helloWorldHash, Size: 11}); err != nil {
		t.Fatal(err)
	}

	if _, ok := d.shouldSkipDownload(context.Background(), "1", "FLAC"); ok {
		t.Error("skipped a track recorded at another quality")
	}
}

// TestShouldSkipDownloadRelocates is the recovery path: the user moved the
// file, and it is recognised by hash and re-recorded where it now lives.
func TestShouldSkipDownloadRelocates(t *testing.T) {
	dir := t.TempDir()
	moved := writeFile(t, filepath.Join(dir, "elsewhere", "renamed.mp3"), "hello world")

	d := newTestDownloader(t, dir)
	stale := filepath.Join(dir, "gone.mp3")
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Quality: "MP3_320", Path: stale, Hash: helloWorldHash, Size: 11}); err != nil {
		t.Fatal(err)
	}

	got, ok := d.shouldSkipDownload(context.Background(), "1", "MP3_320")
	if !ok || got != moved {
		t.Fatalf("shouldSkipDownload() = %q, %v; want %q, true", got, ok, moved)
	}

	stored, err := d.store.DownloadInfo("1")
	if err != nil {
		t.Fatalf("DownloadInfo: %v", err)
	}
	if stored.Path != moved {
		t.Errorf("stored path = %q, want %q", stored.Path, moved)
	}
}

// TestShouldSkipDownloadRelocatesWithoutSize is the same recovery on a ledger
// that has not been migrated: slower, but it must still find the file.
func TestShouldSkipDownloadRelocatesWithoutSize(t *testing.T) {
	dir := t.TempDir()
	moved := writeFile(t, filepath.Join(dir, "elsewhere", "renamed.mp3"), "hello world")

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Quality: "MP3_320", Path: filepath.Join(dir, "gone.mp3"), Hash: helloWorldHash}); err != nil {
		t.Fatal(err)
	}

	got, ok := d.shouldSkipDownload(context.Background(), "1", "MP3_320")
	if !ok || got != moved {
		t.Fatalf("shouldSkipDownload() = %q, %v; want %q, true", got, ok, moved)
	}

	stored, err := d.store.DownloadInfo("1")
	if err != nil {
		t.Fatalf("DownloadInfo: %v", err)
	}
	if stored.Size != int64(len("hello world")) {
		t.Errorf("stored size = %d, want %d: a recovered file should record its size", stored.Size, len("hello world"))
	}
}

func TestShouldSkipDownloadDeletedFile(t *testing.T) {
	dir := t.TempDir()

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Quality: "MP3_320", Path: filepath.Join(dir, "gone.mp3"), Hash: helloWorldHash, Size: 11}); err != nil {
		t.Fatal(err)
	}

	if _, ok := d.shouldSkipDownload(context.Background(), "1", "MP3_320"); ok {
		t.Error("skipped a track whose file no longer exists anywhere")
	}
}

func TestShouldSkipDownloadUnknownTrack(t *testing.T) {
	d := newTestDownloader(t, t.TempDir())

	if _, ok := d.shouldSkipDownload(context.Background(), "missing", "MP3_320"); ok {
		t.Error("skipped a track that was never downloaded")
	}
}
