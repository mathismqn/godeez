package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

// DownloadInfo records one completed download. Quality is stored so that
// re-requesting the same track at a higher quality is not mistaken for a
// duplicate, and Hash lets a file that has since been moved or renamed still
// be recognised.
type DownloadInfo struct {
	TrackID    string    `json:"song_id"`
	Quality    string    `json:"quality"`
	Path       string    `json:"path"`
	Hash       string    `json:"hash"`
	Downloaded time.Time `json:"downloaded_at"`
}

var trackBucket = []byte("tracks")

// DownloadInfo returns the record for trackID. A track that has never been
// downloaded is reported as an error rather than a nil result, and callers
// treat any error the same way: download it.
func (s *Store) DownloadInfo(trackID string) (*DownloadInfo, error) {
	var info DownloadInfo

	if err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(trackBucket)
		if b == nil {
			return errors.New("bucket not found")
		}

		data := b.Get([]byte(trackID))
		if data == nil {
			return errors.New("not found")
		}
		return json.Unmarshal(data, &info)
	}); err != nil {
		return nil, err
	}

	return &info, nil
}

func (s *Store) PutDownloadInfo(d *DownloadInfo) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(trackBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		data, err := json.Marshal(d)
		if err != nil {
			return err
		}

		return b.Put([]byte(d.TrackID), data)
	})
}
