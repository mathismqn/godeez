package deezer

import (
	"encoding/json"
	"path"

	"github.com/flytam/filenamify"
)

type Playlist struct {
	Results struct {
		Data struct {
			Title     string `json:"TITLE"`
			Status    int    `json:"STATUS"`
			CollabKey string `json:"COLLAB_KEY"`
		} `json:"DATA"`
		Songs struct {
			Data []*Song `json:"data"`
		} `json:"SONGS"`
	} `json:"results"`
}

func (p *Playlist) GetType() string {
	return "Playlist"
}

func (p *Playlist) UnmarshalData(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *Playlist) GetSongs() []*Song {
	return p.Results.Songs.Data
}

func (p *Playlist) GetOutputPath(outputDir string) string {
	p.Results.Data.Title, _ = filenamify.Filenamify(p.Results.Data.Title, filenamify.Options{})
	outputPath := path.Join(outputDir, p.Results.Data.Title)
	outputPath, _ = filenamify.Filenamify(outputPath, filenamify.Options{})

	return outputPath
}

func (p *Playlist) GetTitle() string {
	return p.Results.Data.Title
}
