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
			Status   int    `json:"STATUS"`
			Creator  string `json:"PARENT_USERNAME"`
			Duration int    `json:"DURATION"`
		} `json:"DATA"`
		Songs struct {
			Data []*Song `json:"data"`
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
		len(p.Results.Songs.Data),
		time.Duration(p.Results.Data.Duration)*time.Second,
	)
}

func (p *Playlist) GetType() string {
	return "Playlist"
}

func (p *Playlist) GetTitle() string {
	return p.Results.Data.Title
}

func (p *Playlist) GetSongs() []*Song {
	return p.Results.Songs.Data
}

func (p *Playlist) SetSongs(s []*Song) {
	p.Results.Songs.Data = s
}

func (p *Playlist) GetOutputDir(outputDir string) string {
	// For playlists, create a playlist-specific folder for the M3U file
	// The actual songs will be distributed in the tree structure: Artist/Album/Song
	playlistName, _ := filenamify.Filenamify(p.Results.Data.Title, filenamify.Options{})
	return path.Join(outputDir, "Playlists", playlistName)
}

func (p *Playlist) Unmarshal(data []byte) error {
	return json.Unmarshal(data, p)
}
