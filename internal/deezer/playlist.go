package deezer

import (
	"encoding/json"
	"path"

	"github.com/flytam/filenamify"
)

type Playlist struct {
	Data struct {
		Title string `json:"TITLE"`
	} `json:"DATA"`
	Songs struct {
		Data []*Song `json:"data"`
	} `json:"SONGS"`
}

func (p *Playlist) GetURL(id string) string {
	return "https://www.deezer.com/en/playlist/" + id
}

func (p *Playlist) UnmarshalData(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *Playlist) GetSongs() []*Song {
	return p.Songs.Data
}

func (p *Playlist) GetOutputPath(outputDir string) string {
	p.Data.Title, _ = filenamify.Filenamify(p.Data.Title, filenamify.Options{})
	outputPath := path.Join(outputDir, p.Data.Title)
	outputPath, _ = filenamify.Filenamify(outputPath, filenamify.Options{})

	return outputPath
}

func (p *Playlist) GetTitle() string {
	return p.Data.Title
}
