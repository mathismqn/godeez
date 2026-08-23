package download

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mathismqn/godeez/internal/store"
)

// TestMigrateSizesStatsExistingFiles is the common case: every file is where
// the ledger says it is, so the index is never built.
func TestMigrateSizesStatsExistingFiles(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "album", "track.mp3"), "hello world")

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Path: path, Hash: helloWorldHash}); err != nil {
		t.Fatal(err)
	}

	d.migrateSizes(context.Background())

	got, err := d.store.DownloadInfo("1")
	if err != nil {
		t.Fatalf("DownloadInfo: %v", err)
	}
	if got.Size != int64(len("hello world")) {
		t.Errorf("Size = %d, want %d", got.Size, len("hello world"))
	}
	if d.fileIndex != nil {
		t.Error("built the file index for a ledger where nothing had moved")
	}
	if v := d.store.Version(); v != sizeSchema {
		t.Errorf("Version() = %d, want %d", v, sizeSchema)
	}
}

// TestMigrateSizesRepairsMovedFile covers the escalation: a record whose path
// is dead comes out with both its path and its size corrected.
func TestMigrateSizesRepairsMovedFile(t *testing.T) {
	dir := t.TempDir()
	moved := writeFile(t, filepath.Join(dir, "elsewhere", "renamed.mp3"), "hello world")

	d := newTestDownloader(t, dir)
	stale := filepath.Join(dir, "album", "track.mp3")
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Path: stale, Hash: helloWorldHash}); err != nil {
		t.Fatal(err)
	}

	d.migrateSizes(context.Background())

	got, err := d.store.DownloadInfo("1")
	if err != nil {
		t.Fatalf("DownloadInfo: %v", err)
	}
	if got.Path != moved {
		t.Errorf("Path = %q, want %q", got.Path, moved)
	}
	if got.Size != int64(len("hello world")) {
		t.Errorf("Size = %d, want %d", got.Size, len("hello world"))
	}
}

// TestMigrateSizesRunsOnce is what the version marker is for: a record that
// could not be filled in must not make every later run scan the library.
func TestMigrateSizesRunsOnce(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "album", "track.mp3"), "hello world")

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Path: path, Hash: helloWorldHash}); err != nil {
		t.Fatal(err)
	}

	d.migrateSizes(context.Background())

	if v := d.store.Version(); v != sizeSchema {
		t.Fatalf("Version() = %d, want %d", v, sizeSchema)
	}

	second := &Downloader{appConfig: d.appConfig, store: d.store}
	second.migrateSizes(context.Background())
	if second.fileIndex != nil {
		t.Error("scanned the library again on an already migrated ledger")
	}
}

// TestMigrateSizesDropsDeletedFile covers a record the pass can prove is
// dead: nothing at its path, and nothing matching its hash in a library that
// was read in full.
func TestMigrateSizesDropsDeletedFile(t *testing.T) {
	dir := t.TempDir()
	kept := writeFile(t, filepath.Join(dir, "album", "track.mp3"), "hello world")

	d := newTestDownloader(t, dir)
	if err := d.store.UpdateDownloadInfos([]*store.DownloadInfo{
		{TrackID: "1", Path: kept, Hash: helloWorldHash},
		{TrackID: "2", Path: filepath.Join(dir, "album", "deleted.mp3"), Hash: "deadbeef"},
	}, nil); err != nil {
		t.Fatal(err)
	}

	d.migrateSizes(context.Background())

	if _, err := d.store.DownloadInfo("2"); err == nil {
		t.Error("kept a record whose file is gone from a library that was read")
	}
	if _, err := d.store.DownloadInfo("1"); err != nil {
		t.Errorf("dropped the record of a file that is still there: %v", err)
	}
}

// TestMigrateSizesKeepsRecordsWhenLibraryEmpty is one half of the guard on
// that deletion: a walk that finds no audio at all cannot tell a library the
// user emptied from one that is not where it should be.
func TestMigrateSizesKeepsRecordsWhenLibraryEmpty(t *testing.T) {
	dir := t.TempDir()

	d := newTestDownloader(t, dir)
	gone := filepath.Join(dir, "album", "deleted.mp3")
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Path: gone, Hash: helloWorldHash}); err != nil {
		t.Fatal(err)
	}

	d.migrateSizes(context.Background())

	got, err := d.store.DownloadInfo("1")
	if err != nil {
		t.Fatalf("dropped a record although the library held no audio at all: %v", err)
	}
	if got.Size != 0 {
		t.Errorf("Size = %d, want 0 for a file that could not be found", got.Size)
	}
}

// TestMigrateSizesCanceled must not record the migration as done, so the run
// that follows finishes the job.
func TestMigrateSizesCanceled(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "track.mp3"), "hello world")

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Path: path, Hash: helloWorldHash}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d.migrateSizes(ctx)

	if v := d.store.Version(); v != 0 {
		t.Errorf("Version() = %d, want 0 after a canceled migration", v)
	}
}

// TestMigrateSizesLeavesRecordedSizes makes sure the pass is not undoing the
// work of the downloader, which records a size of its own.
func TestMigrateSizesLeavesRecordedSizes(t *testing.T) {
	dir := t.TempDir()

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{TrackID: "1", Path: "/gone/track.mp3", Hash: helloWorldHash, Size: 42}); err != nil {
		t.Fatal(err)
	}

	d.migrateSizes(context.Background())

	got, err := d.store.DownloadInfo("1")
	if err != nil {
		t.Fatalf("DownloadInfo: %v", err)
	}
	if got.Size != 42 || got.Path != "/gone/track.mp3" {
		t.Errorf("record = %+v, want it left alone", got)
	}
	if d.fileIndex != nil {
		t.Error("built the file index for a record that already had a size")
	}
}

// TestMigrateSizesKeepsRecordsWhenLibraryPartlyUnreadable is the other half,
// and the one an emptiness check alone would miss: the walk looks trustworthy
// but a file inside the part it could not read is out of sight, not gone.
func TestMigrateSizesKeepsRecordsWhenLibraryPartlyUnreadable(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "album", "track.mp3"), "hello world")

	locked := filepath.Join(dir, "locked")
	if err := os.MkdirAll(locked, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "hidden.mp3"), []byte("out of sight"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0755) })

	d := newTestDownloader(t, dir)
	if err := d.store.PutDownloadInfo(&store.DownloadInfo{
		TrackID: "1",
		Path:    filepath.Join(locked, "hidden.mp3"),
		Hash:    "0000000000000000000000000000000000000000000000000000000000000000",
	}); err != nil {
		t.Fatal(err)
	}

	d.migrateSizes(context.Background())

	if !d.fileIndex.degraded {
		t.Fatal("index reported a complete walk although a directory could not be read")
	}
	if _, err := d.store.DownloadInfo("1"); err != nil {
		t.Errorf("dropped a record whose file was merely out of reach: %v", err)
	}
}
