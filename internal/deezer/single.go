package deezer

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"time"
)

type Single struct {
	Results struct {
		Data *Track `json:"DATA"`
	} `json:"results"`
}

func (s *Single) String() string {
	if s.Results.Data == nil {
		return "Track: No data available"
	}

	duration, err := strconv.Atoi(s.Results.Data.Duration)
	if err != nil {
		duration = 0
	}

	return fmt.Sprintf(
		`================= [ Track Info ] =================
Title:    %s
Artist:   %s
Duration: %s
==================================================`,
		s.Results.Data.GetTitle(),
		s.Results.Data.Artist,
		time.Duration(duration)*time.Second,
	)
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
