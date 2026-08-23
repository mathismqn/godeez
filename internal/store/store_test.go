package store

import (
	"encoding/json"
	"errors"
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

// TestLegacyRecordHasNoSize pins the meaning of a zero size: records written
// before the field existed decode without one, and readers rely on that to
// fall back to hashing.
func TestLegacyRecordHasNoSize(t *testing.T) {
	var info DownloadInfo
	legacy := `{"song_id":"1","quality":"FLAC","path":"/music/a.flac","hash":"abc","downloaded_at":"2024-01-01T00:00:00Z"}`

	if err := json.Unmarshal([]byte(legacy), &info); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if info.Size != 0 {
		t.Errorf("Size = %d, want 0", info.Size)
	}
	if info.Hash != "abc" {
		t.Errorf("Hash = %q, want %q", info.Hash, "abc")
	}
}

func TestSizeRoundTrip(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if err := s.PutDownloadInfo(&DownloadInfo{TrackID: "1", Size: 4096}); err != nil {
		t.Fatalf("PutDownloadInfo: %v", err)
	}

	got, err := s.DownloadInfo("1")
	if err != nil {
		t.Fatalf("DownloadInfo: %v", err)
	}
	if got.Size != 4096 {
		t.Errorf("Size = %d, want 4096", got.Size)
	}
}

func TestEachDownloadInfo(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	// An empty ledger is a legitimate no-op rather than an error.
	seen := 0
	if err := s.EachDownloadInfo(func(*DownloadInfo) error { seen++; return nil }); err != nil {
		t.Fatalf("EachDownloadInfo on an empty store: %v", err)
	}
	if seen != 0 {
		t.Errorf("visited %d records of an empty store", seen)
	}

	want := []*DownloadInfo{
		{TrackID: "1", Size: 10},
		{TrackID: "2", Size: 20},
	}
	if err := s.UpdateDownloadInfos(want, nil); err != nil {
		t.Fatalf("UpdateDownloadInfos: %v", err)
	}

	sizes := map[string]int64{}
	if err := s.EachDownloadInfo(func(info *DownloadInfo) error {
		sizes[info.TrackID] = info.Size
		return nil
	}); err != nil {
		t.Fatalf("EachDownloadInfo: %v", err)
	}

	if len(sizes) != 2 || sizes["1"] != 10 || sizes["2"] != 20 {
		t.Errorf("EachDownloadInfo visited %v, want {1:10 2:20}", sizes)
	}
}

func TestEachDownloadInfoStops(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if err := s.UpdateDownloadInfos([]*DownloadInfo{{TrackID: "1"}, {TrackID: "2"}}, nil); err != nil {
		t.Fatalf("UpdateDownloadInfos: %v", err)
	}

	stop := errors.New("stop")
	if err := s.EachDownloadInfo(func(*DownloadInfo) error { return stop }); !errors.Is(err, stop) {
		t.Errorf("EachDownloadInfo() = %v, want %v", err, stop)
	}
}

func TestVersion(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if v := s.Version(); v != 0 {
		t.Errorf("Version() on a fresh store = %d, want 0", v)
	}

	if err := s.SetVersion(3); err != nil {
		t.Fatalf("SetVersion: %v", err)
	}
	if v := s.Version(); v != 3 {
		t.Errorf("Version() = %d, want 3", v)
	}
}

// TestUpdateDownloadInfosRemoves covers the migration's other half: dropping
// records in the same transaction that sizes the ones that remain.
func TestUpdateDownloadInfosRemoves(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if err := s.UpdateDownloadInfos([]*DownloadInfo{{TrackID: "1"}, {TrackID: "2"}}, nil); err != nil {
		t.Fatalf("UpdateDownloadInfos: %v", err)
	}

	if err := s.UpdateDownloadInfos([]*DownloadInfo{{TrackID: "1", Size: 10}}, []string{"2"}); err != nil {
		t.Fatalf("UpdateDownloadInfos: %v", err)
	}

	got, err := s.DownloadInfo("1")
	if err != nil {
		t.Fatalf("DownloadInfo: %v", err)
	}
	if got.Size != 10 {
		t.Errorf("Size = %d, want 10", got.Size)
	}

	if _, err := s.DownloadInfo("2"); err == nil {
		t.Error("expected the removed record to be gone")
	}

	// Removing an id that is not there is not an error: the migration does
	// not check first.
	if err := s.UpdateDownloadInfos(nil, []string{"missing"}); err != nil {
		t.Errorf("UpdateDownloadInfos on an unknown id: %v", err)
	}
}
