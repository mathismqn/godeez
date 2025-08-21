package deezer

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/flytam/filenamify"
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

	duration := "Unknown"
	if t.Results.Data.Duration != "" {
		if d, err := time.ParseDuration(t.Results.Data.Duration + "s"); err == nil {
			duration = d.String()
		}
	}

	return fmt.Sprintf(
		`================= [ Track Info ] =================
Title:    %s
Artist:   %s
Duration: %s
==================================================`,
		t.Results.Data.GetTitle(),
		t.Results.Data.Artist,
		duration,
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
		return []*Song{}
	}
	return []*Song{t.Results.Data}
}

func (t *Track) SetSongs(songs []*Song) {
	if len(songs) > 0 {
		t.Results.Data = songs[0]
	}
}

func (t *Track) GetOutputDir(outputDir string) string {
	if t.Results.Data == nil {
		return outputDir
	}

	// For single tracks, create a simple "Singles" folder
	base := "Singles"
	base, _ = filenamify.Filenamify(base, filenamify.Options{})
	outputDir = path.Join(outputDir, base)

	return outputDir
}

func (t *Track) Unmarshal(data []byte) error {
	return json.Unmarshal(data, t)
}
