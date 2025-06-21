package deezer

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/flytam/filenamify"
)

type Artist struct {
	Results struct {
		Data struct {
			Name string `json:"ART_NAME"`
		} `json:"DATA"`
		Songs struct {
			Data []*Song `json:"data"`
		} `json:"TOP"`
	} `json:"results"`
}

func (a *Artist) GetType() string {
	return "Artist"
}

func (a *Artist) GetTitle() string {
	return a.Results.Data.Name
}

func (a *Artist) GetSongs() []*Song {
	return a.Results.Songs.Data
}

func (a *Artist) SetSongs(s []*Song) {
	a.Results.Songs.Data = s
}

func (a *Artist) GetOutputDir(outputDir string) string {
	base, _ := filenamify.Filenamify(a.GetTitle(), filenamify.Options{})
	return path.Join(outputDir, base)
}

func (a *Artist) Unmarshal(data []byte) error {
	return json.Unmarshal(data, a)
}

func (a *Artist) String() string {
	tracks := a.GetSongs()
	count := len(tracks)

	limit := 3
	if count < limit {
		limit = count
	}

	totalSec := 0
	for _, s := range tracks {
		if d, err := strconv.Atoi(s.Duration); err == nil {
			totalSec += d
		}
	}
	totalDuration := time.Duration(totalSec) * time.Second

	var b strings.Builder
	fmt.Fprintf(&b, "============= [ Artist Info ] =============\n")
	fmt.Fprintf(&b, "Artist:   %s\n", a.GetTitle())
	fmt.Fprintf(&b, "Tracks:   %d\n", count)
	fmt.Fprintf(&b, "Playtime: %s\n", totalDuration)
	fmt.Fprintf(&b, "-------------------------------------------\n")
	fmt.Fprintf(&b, "Top %d most popular tracks:\n", limit)
	for i := 0; i < limit; i++ {
		s := tracks[i]
		title := s.GetTitle()
		fmt.Fprintf(&b, "    %2d. %s – %s\n", i+1, s.Artist, title)
	}
	fmt.Fprintf(&b, "===========================================\n")

	return b.String()
}
