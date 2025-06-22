package store

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

type WatchedPlaylist struct {
	ID      string        `json:"id"`
	Quality string        `json:"quality"`
	BPM     bool          `json:"bpm"`
	Timeout time.Duration `json:"timeout"`
}

var watchedBucket = []byte("watched")

func ListWatchedPlaylists() ([]*WatchedPlaylist, error) {
	var playlists []*WatchedPlaylist
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(watchedBucket)
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var p WatchedPlaylist
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			playlists = append(playlists, &p)

			return nil
		})
	}); err != nil {
		return nil, err
	}

	return playlists, nil
}

func (p *WatchedPlaylist) Save() error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(watchedBucket)
		if err != nil {
			return err
		}

		data, err := json.Marshal(p)
		if err != nil {
			return err
		}

		return b.Put([]byte(p.ID), data)
	})
}

func RemoveWatchedPlaylist(playlistID string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(watchedBucket)
		if b == nil {
			return fmt.Errorf("bucket not found")
		}

		return b.Delete([]byte(playlistID))
	})
}

func IsWatched(playlistID string) (bool, error) {
	var found bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(watchedBucket)
		if b == nil {
			return nil
		}
		found = b.Get([]byte(playlistID)) != nil

		return nil
	})

	return found, err
}
