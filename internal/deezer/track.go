package deezer

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"time"
)

type Track struct {
	Results struct {
		Data *Song `json:"DATA"`
	} `json:"results"`
}

func (t *Track) String() string {
	if t.Results.Data == nil {
		return "Track: No data available"
	}

	duration, err := strconv.Atoi(t.Results.Data.Duration)
	if err != nil {
		duration = 0
	}

	return fmt.Sprintf(
		`================= [ Track Info ] =================
Title:    %s
Artist:   %s
Duration: %s
==================================================`,
		t.Results.Data.GetTitle(),
		t.Results.Data.Artist,
		time.Duration(duration)*time.Second,
	)
}

func (t *Track) GetType() string {
	return "Track"
}

func (t *Track) GetTitle() string {
	if t.Results.Data == nil {
		return ""
	}
	return t.Results.Data.GetTitle()
}

func (t *Track) GetSongs() []*Song {
	if t.Results.Data == nil {
		return nil
	}
	return []*Song{t.Results.Data}
}

func (t *Track) SetSongs(songs []*Song) {}

func (t *Track) GetOutputDir(outputDir string) string {
	return path.Join(outputDir, "Singles")
}

func (t *Track) Unmarshal(data []byte) error {
	return json.Unmarshal(data, t)
}
