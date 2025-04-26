package db

import (
	"fmt"
	"os"
	"path"

	bolt "go.etcd.io/bbolt"
)

var (
	db          *bolt.DB
	trackBucket = []byte("tracks")
)

func Init(cfgDir string) {
	var err error

	dbPath := path.Join(cfgDir, "tracks.db")
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
