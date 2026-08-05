package deezer

import (
	"encoding/json"
	"path"
)

type Single struct {
	Results struct {
		Data *Track `json:"DATA"`
	} `json:"results"`
}

func (s *Single) GetTitle() string {
	if s.Results.Data == nil {
		return ""
	}
	return s.Results.Data.GetTitle()
}

func (s *Single) GetTracks() []*Track {
	if s.Results.Data == nil {
		return nil
	}
	return []*Track{s.Results.Data}
}

func (s *Single) SetTracks(tracks []*Track) {}

func (s *Single) GetOutputDir(outputDir string) string {
	return path.Join(outputDir, "Singles")
}

func (s *Single) Unmarshal(data []byte) error {
	return json.Unmarshal(data, s)
}
