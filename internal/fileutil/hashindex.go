package fileutil

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type HashIndex struct {
	files map[string]string
}

func NewHashIndex(ctx context.Context, root string) (*HashIndex, error) {
	index := &HashIndex{files: make(map[string]string)}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil || d.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			return nil
		}

		index.files[hex.EncodeToString(h.Sum(nil))] = path
		return nil
	})
	if err != nil {
		return nil, err
	}

	return index, nil
}

func (h *HashIndex) Find(hash string) (string, bool) {
	path, ok := h.files[hash]
	return path, ok
}
