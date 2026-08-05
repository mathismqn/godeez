package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenPutGet(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if _, err := os.Stat(filepath.Join(dir, ".tracks.db")); err != nil {
		t.Errorf("expected .tracks.db to exist: %v", err)
	}

	info := &DownloadInfo{
		TrackID:    "123",
		Quality:    "MP3_320",
		Path:       "/music/track.mp3",
		Hash:       "abc",
		Downloaded: time.Now().Truncate(time.Second),
	}
	if err := s.PutDownloadInfo(info); err != nil {
		t.Fatalf("PutDownloadInfo: %v", err)
	}

	got, err := s.DownloadInfo("123")
	if err != nil {
		t.Fatalf("DownloadInfo: %v", err)
	}
	if got.TrackID != info.TrackID || got.Quality != info.Quality || got.Path != info.Path || got.Hash != info.Hash {
		t.Errorf("DownloadInfo() = %+v, want %+v", got, info)
	}
}

func TestDownloadInfoNotFound(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if _, err := s.DownloadInfo("missing"); err == nil {
		t.Error("expected error for unknown track ID")
	}
}

func TestClose(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
