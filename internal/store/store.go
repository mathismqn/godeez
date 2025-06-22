package store

import (
	"fmt"
	"path"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

func OpenDB(cfgDir string) error {
	var err error

	dbPath := path.Join(cfgDir, "tracks.db")
	db, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	return nil
}
