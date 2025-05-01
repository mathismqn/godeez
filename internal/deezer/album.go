package deezer

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/flytam/filenamify"
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
		} `json:"DATA"`
		Songs struct {
			Data []*Song `json:"data"`
		} `json:"SONGS"`
	} `json:"results"`
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

func (a *Album) GetOutputDir(outputDir string) string {
	base := fmt.Sprintf("%s - %s", a.Results.Data.Artist, a.Results.Data.Title)
	base, _ = filenamify.Filenamify(base, filenamify.Options{})
	outputDir = path.Join(outputDir, base)

	return outputDir
}

func (a *Album) Unmarshal(data []byte) error {
	return json.Unmarshal(data, a)
}
