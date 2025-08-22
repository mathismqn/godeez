package store

import (
	"fmt"
	"os"
	"path"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

func OpenDB(cfgDir string) error {
	var err error

	// Ensure the directory exists
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	dbPath := path.Join(cfgDir, "tracks.db")
	db, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	return nil
}

// GetDB returns the database connection
func GetDB() (*bolt.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized, call OpenDB first")
	}
	return db, nil
}
