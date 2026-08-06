// Package store keeps the ledger of what has already been downloaded, so a
// repeated run can skip tracks instead of fetching them again.
//
// It is a bbolt database written as a hidden file inside the output
// directory, which keeps it travelling with the music library it describes.
// Losing it is harmless: the worst outcome is re-downloading.
package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
	bolterrors "go.etcd.io/bbolt/errors"
)

type Store struct {
	db *bbolt.DB
}

// Open opens the ledger in dir, creating it if needed.
//
// bbolt takes an exclusive file lock, so a second godeez running against the
// same output directory blocks here. The timeout turns that into a clear
// message rather than an apparent hang.
func Open(dir string) (*Store, error) {
	db, err := bbolt.Open(filepath.Join(dir, ".tracks.db"), 0600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		if errors.Is(err, bolterrors.ErrTimeout) {
			return nil, errors.New("database is already in use by another process")
		}
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
