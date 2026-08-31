package download

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/store"
)

// helloWorldHash is the sha256 of "hello world", the content every fixture
// file in this package is written with.
const helloWorldHash = "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

// writeFile creates path, including its parent directories, and returns it.
func writeFile(t *testing.T, path, content string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	return path
}

// discLayoutOf builds a layout from one disc label per track.
func discLayoutOf(discs ...string) discLayout {
	tracks := make([]*deezer.Track, len(discs))
	for i, disc := range discs {
		tracks[i] = &deezer.Track{DiscNumber: deezer.Number(disc)}
	}

	return newDiscLayout(tracks)
}

// newTestDownloader builds a Downloader whose library and ledger are both dir.
func newTestDownloader(t *testing.T, dir string) *Downloader {
	t.Helper()

	st, err := store.Open(dir)
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	return &Downloader{appConfig: &config.Config{OutputDir: dir}, store: st}
}
