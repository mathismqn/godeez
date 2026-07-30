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
		Tracks struct {
			Data []*Track `json:"data"`
		} `json:"TOP"`
	} `json:"results"`
}

func (a *Artist) String() string {
	tracks := a.Results.Tracks.Data
	count := len(tracks)

	totalSec := 0
	for _, t := range tracks {
		if d, err := strconv.Atoi(t.Duration); err == nil {
			totalSec += d
		}
	}

	limit := min(3, count)

	var b strings.Builder
	fmt.Fprintf(&b, "============= [ Artist Info ] =============\n")
	fmt.Fprintf(&b, "Artist:   %s\n", a.Results.Data.Name)
	fmt.Fprintf(&b, "Tracks:   %d\n", count)
	fmt.Fprintf(&b, "Playtime: %s\n", time.Duration(totalSec)*time.Second)
	fmt.Fprintf(&b, "-------------------------------------------\n")
	fmt.Fprintf(&b, "Top %d most popular tracks:\n", limit)
	for i := 0; i < limit; i++ {
		t := tracks[i]
		fmt.Fprintf(&b, "    %2d. %s – %s\n", i+1, t.Artist, t.GetTitle())
	}
	fmt.Fprintf(&b, "===========================================\n")

	return b.String()
}

func (a *Artist) GetType() string {
	return "Artist"
}

func (a *Artist) GetTitle() string {
	return a.Results.Data.Name
}

func (a *Artist) GetTracks() []*Track {
	return a.Results.Tracks.Data
}

func (a *Artist) SetTracks(t []*Track) {
	a.Results.Tracks.Data = t
}

func (a *Artist) GetOutputDir(outputDir string) string {
	base, _ := filenamify.Filenamify(a.Results.Data.Name, filenamify.Options{})
	return path.Join(outputDir, base)
}

func (a *Artist) Unmarshal(data []byte) error {
	return json.Unmarshal(data, a)
}
