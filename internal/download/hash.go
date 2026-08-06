package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// hashIndex maps content hash to path for everything under the output
// directory. It is what lets a moved or renamed file still be recognised as
// an existing download.
type hashIndex struct {
	files map[string]string
}

// newHashIndex hashes every file under root.
//
// Unreadable files and directories are skipped rather than failing the walk,
// since a permission error somewhere in a music library should not break the
// skip check. Only cancellation aborts it. Duplicated content collapses to
// whichever path is walked last, which is fine: any copy is a valid answer.
func newHashIndex(ctx context.Context, root string) (*hashIndex, error) {
	index := &hashIndex{files: make(map[string]string)}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil || d.IsDir() {
			return nil
		}

		hash, err := hashFile(path)
		if err != nil {
			return nil
		}

		index.files[hash] = path
		return nil
	})
	if err != nil {
		return nil, err
	}

	return index, nil
}

func (h *hashIndex) find(hash string) (string, bool) {
	path, ok := h.files[hash]
	return path, ok
}

// initHashIndex builds the index on first use and reuses it afterwards.
//
// Building it means hashing an entire music library, so it is deferred until
// something actually needs it: a run where every recorded path is still valid
// never pays that cost. The error is cached alongside the index so a failed
// build is not retried once per track.
func (d *Downloader) initHashIndex(ctx context.Context) error {
	d.hashIndexOnce.Do(func() {
		d.hashIndex, d.hashIndexErr = newHashIndex(ctx, d.appConfig.OutputDir)
	})

	return d.hashIndexErr
}
