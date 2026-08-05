package deezer

import (
	"encoding/json"
	"path"

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
