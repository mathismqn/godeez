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
	"strconv"
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

var metaBucket = []byte("meta")
var versionKey = []byte("version")

// Version reports the version stamped on the ledger, which is 0 until a
// migration sets one. The store only keeps the number; what it means belongs
// to whichever migration stamped it. An unreadable marker reports 0, since a
// redundant migration is cheaper than a skipped one.
func (s *Store) Version() int {
	version := 0

	_ = s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(metaBucket)
		if b == nil {
			return nil
		}

		if v, err := strconv.Atoi(string(b.Get(versionKey))); err == nil {
			version = v
		}
		return nil
	})

	return version
}

func (s *Store) SetVersion(version int) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(metaBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		return b.Put(versionKey, []byte(strconv.Itoa(version)))
	})
}
