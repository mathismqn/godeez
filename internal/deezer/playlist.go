package deezer

import (
	"encoding/json"
	"path"
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
	return path.Join(outputDir, p.Data.Title)
}

func (p *Playlist) GetTitle() string {
	return p.Data.Title
}
