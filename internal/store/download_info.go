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
// be recognised. Size makes that lookup cheap by rejecting a candidate with a
// stat instead of a full read; zero means unknown, and every reader falls
// back to hashing.
type DownloadInfo struct {
	TrackID    string    `json:"song_id"`
	Quality    string    `json:"quality"`
	Path       string    `json:"path"`
	Hash       string    `json:"hash"`
	Size       int64     `json:"size"`
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

// EachDownloadInfo calls fn for every record in the ledger. An empty ledger
// is a no-op rather than an error, and an entry that fails to decode is
// skipped rather than costing the caller the rest.
//
// fn runs inside a read transaction and so must not write to the store.
func (s *Store) EachDownloadInfo(fn func(*DownloadInfo) error) error {
	return s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(trackBucket)
		if b == nil {
			return nil
		}

		return b.ForEach(func(_, data []byte) error {
			var info DownloadInfo
			if json.Unmarshal(data, &info) != nil {
				return nil
			}
			return fn(&info)
		})
	})
}

func (s *Store) PutDownloadInfo(d *DownloadInfo) error {
	return s.UpdateDownloadInfos([]*DownloadInfo{d}, nil)
}

// UpdateDownloadInfos writes put and removes remove in one transaction, which
// costs a single fsync instead of one per track.
//
// Both halves committing together matters for the migration that uses it: it
// never runs again, so a half applied result would be permanent.
func (s *Store) UpdateDownloadInfos(put []*DownloadInfo, remove []string) error {
	if len(put) == 0 && len(remove) == 0 {
		return nil
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(trackBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		for _, d := range put {
			data, err := json.Marshal(d)
			if err != nil {
				return err
			}

			if err := b.Put([]byte(d.TrackID), data); err != nil {
				return err
			}
		}

		for _, trackID := range remove {
			if err := b.Delete([]byte(trackID)); err != nil {
				return err
			}
		}

		return nil
	})
}
