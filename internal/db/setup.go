package db

import (
	"fmt"
	"os"
	"path"

	"github.com/mathismqn/godeez/internal/utils"
	bolt "go.etcd.io/bbolt"
)

var (
	db          *bolt.DB
	trackBucket = []byte("tracks")
)

func Init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not get home directory: %v\n", err)
		os.Exit(1)
	}

	appDir := path.Join(homeDir, ".godeez")
	if err := utils.EnsureDir(appDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create app directory: %v\n", err)
		os.Exit(1)
	}

	dbPath := path.Join(appDir, "tracks.db")
	db, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not open database: %v\n", err)
		os.Exit(1)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(trackBucket)
		return err
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create bucket: %v\n", err)
		os.Exit(1)
	}
}
