package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

type DownloadInfo struct {
	TrackID    string    `json:"song_id"`
	Quality    string    `json:"quality"`
	Path       string    `json:"path"`
	Hash       string    `json:"hash"`
	Downloaded time.Time `json:"downloaded_at"`
}

var trackBucket = []byte("tracks")

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
