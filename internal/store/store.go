package store

import (
	"fmt"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

type Store struct {
	db *bolt.DB
}

func Open(dir string) (*Store, error) {
	db, err := bolt.Open(filepath.Join(dir, ".tracks.db"), 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
