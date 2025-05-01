package store

import (
	"encoding/json"
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

type DownloadInfo struct {
	SongID     string    `json:"song_id"`
	Quality    string    `json:"quality"`
	Path       string    `json:"path"`
	Hash       string    `json:"hash"`
	Downloaded time.Time `json:"downloaded_at"`
}

func GetDownloadInfo(songID string) (*DownloadInfo, error) {
	var info DownloadInfo

	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(trackBucket)
		if b == nil {
			return fmt.Errorf("bucket not found")
		}

		data := b.Get([]byte(songID))
		if data == nil {
			return fmt.Errorf("not found")
		}
		return json.Unmarshal(data, &info)
	}); err != nil {
		return nil, err
	}

	return &info, nil
}

func (d *DownloadInfo) Save() error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(trackBucket)
		if b == nil {
			return fmt.Errorf("bucket not found")
		}

		data, err := json.Marshal(d)
		if err != nil {
			return err
		}

		err = b.Put([]byte(d.SongID), data)
		if err != nil {
			return err
		}

		return nil
	})
}
