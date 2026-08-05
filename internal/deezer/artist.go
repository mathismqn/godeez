package deezer

import (
	"encoding/json"
	"path/filepath"

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

func (a *Artist) Title() string {
	return a.Results.Data.Name
}

func (a *Artist) Tracks() []*Track {
	return a.Results.Tracks.Data
}

func (a *Artist) SetTracks(t []*Track) {
	a.Results.Tracks.Data = t
}

func (a *Artist) OutputDir(outputDir string) string {
	base, _ := filenamify.Filenamify(a.Results.Data.Name, filenamify.Options{})
	return filepath.Join(outputDir, base)
}

func (a *Artist) decode(data []byte) error {
	return json.Unmarshal(data, a)
}
