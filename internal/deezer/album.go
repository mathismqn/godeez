package deezer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Album struct {
	Results struct {
		Data struct {
			Title               string `json:"ALB_TITLE"`
			Artist              string `json:"ART_NAME"`
			OriginalReleaseDate string `json:"ORIGINAL_RELEASE_DATE"`
			PhysicalReleaseDate string `json:"PHYSICAL_RELEASE_DATE"`
			Label               string `json:"LABEL_NAME"`
			ProducerLine        string `json:"PRODUCER_LINE"`
			Duration            string `json:"DURATION"`
		} `json:"DATA"`
		Songs struct {
			Data []*Song `json:"data"`
		} `json:"SONGS"`
	} `json:"results"`
}

func (a *Album) String() string {
	duration, err := strconv.Atoi(a.Results.Data.Duration)
	if err != nil {
		duration = 0
	}

	return fmt.Sprintf(
		`================= [ Album Info ] =================
Title:    %s
Artist:   %s
Tracks:   %d
Duration: %s
==================================================`,
		a.Results.Data.Title,
		a.Results.Data.Artist,
		len(a.Results.Songs.Data),
		time.Duration(duration)*time.Second,
	)
}

func (a *Album) GetType() string {
	return "Album"
}

func (a *Album) GetTitle() string {
	return a.Results.Data.Title
}

func (a *Album) GetSongs() []*Song {
	return a.Results.Songs.Data
}

func (a *Album) SetSongs(s []*Song) {
	a.Results.Songs.Data = s
}

func (a *Album) GetOutputDir(outputDir string) string {
	// For albums, return the base output directory since songs will handle their own paths
	// The actual song paths will be: Artist/Album/Song
	return outputDir
}

func (a *Album) Unmarshal(data []byte) error {
	return json.Unmarshal(data, a)
}
