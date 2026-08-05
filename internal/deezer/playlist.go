package deezer

import (
	"encoding/json"
	"path"

	"github.com/flytam/filenamify"
)

type Playlist struct {
	Results struct {
		Data struct {
			Title    string `json:"TITLE"`
			Creator  string `json:"PARENT_USERNAME"`
			Duration int    `json:"DURATION"`
		} `json:"DATA"`
		Tracks struct {
			Data []*Track `json:"data"`
		} `json:"SONGS"`
	} `json:"results"`
}

func (p *Playlist) Title() string {
	return p.Results.Data.Title
}

func (p *Playlist) Tracks() []*Track {
	return p.Results.Tracks.Data
}

func (p *Playlist) SetTracks(t []*Track) {
	p.Results.Tracks.Data = t
}

func (p *Playlist) OutputDir(outputDir string) string {
	base, _ := filenamify.Filenamify(p.Results.Data.Title, filenamify.Options{})
	return path.Join(outputDir, base)
}

func (p *Playlist) decode(data []byte) error {
	return json.Unmarshal(data, p)
}
