package deezer

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/flytam/filenamify"
)

type Album struct {
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
}

func (a *Album) GetURL(id string) string {
	return "https://www.deezer.com/en/album/" + id
}

func (a *Album) UnmarshalData(data []byte) error {
	return json.Unmarshal(data, a)
}

func (a *Album) GetSongs() []*Song {
	return a.Songs.Data
}

func (a *Album) GetOutputPath(outputDir string) string {
	base := fmt.Sprintf("%s - %s", a.Data.Artist, a.Data.Title)
	base, _ = filenamify.Filenamify(base, filenamify.Options{})
	outputPath := path.Join(outputDir, base)
	outputPath, _ = filenamify.Filenamify(outputPath, filenamify.Options{})

	return outputPath
}

func (a *Album) GetTitle() string {
	return a.Data.Title
}
