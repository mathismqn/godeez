package deezer

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

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

func (p *Playlist) String() string {
	return fmt.Sprintf(
		`=============== [ Playlist Info ] ===============
Title:    %s
Creator:  %s
Tracks:   %d
Duration: %s
=================================================`,
		p.Results.Data.Title,
		p.Results.Data.Creator,
		len(p.Results.Tracks.Data),
		time.Duration(p.Results.Data.Duration)*time.Second,
	)
}

func (p *Playlist) GetTitle() string {
	return p.Results.Data.Title
}

func (p *Playlist) GetTracks() []*Track {
	return p.Results.Tracks.Data
}

func (p *Playlist) SetTracks(t []*Track) {
	p.Results.Tracks.Data = t
}

func (p *Playlist) GetOutputDir(outputDir string) string {
	base, _ := filenamify.Filenamify(p.Results.Data.Title, filenamify.Options{})
	return path.Join(outputDir, base)
}

func (p *Playlist) Unmarshal(data []byte) error {
	return json.Unmarshal(data, p)
}
