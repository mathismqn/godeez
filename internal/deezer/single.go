package deezer

import (
	"encoding/json"
	"path/filepath"
)

type Single struct {
	Results struct {
		Data *Track `json:"DATA"`
	} `json:"results"`
}

func (s *Single) Title() string {
	if s.Results.Data == nil {
		return ""
	}
	return s.Results.Data.FullTitle()
}

func (s *Single) Tracks() []*Track {
	if s.Results.Data == nil {
		return nil
	}
	return []*Track{s.Results.Data}
}

func (s *Single) SetTracks(tracks []*Track) {}

func (s *Single) OutputDir(outputDir string) string {
	return filepath.Join(outputDir, "Singles")
}

func (s *Single) decode(data []byte) error {
	return json.Unmarshal(data, s)
}
