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
