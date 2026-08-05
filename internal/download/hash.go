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

type hashIndex struct {
	files map[string]string
}

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

func (d *Downloader) initHashIndex(ctx context.Context) error {
	d.hashIndexOnce.Do(func() {
		d.hashIndex, d.hashIndexErr = newHashIndex(ctx, d.appConfig.OutputDir)
	})

	return d.hashIndexErr
}
