package deezer

import (
	"fmt"

	"github.com/flytam/filenamify"
)

type Song struct {
	ID           string `json:"SNG_ID"`
	Artist       string `json:"ART_NAME"`
	Title        string `json:"SNG_TITLE"`
	Version      string `json:"VERSION"`
	Cover        string `json:"ALB_PICTURE"`
	Contributors struct {
		MainArtists []string `json:"main_artist"`
		Composers   []string `json:"composer"`
		Authors     []string `json:"author"`
	} `json:"SNG_CONTRIBUTORS"`
	Duration    string `json:"DURATION"`
	Gain        string `json:"GAIN"`
	ISRC        string `json:"ISRC"`
	TrackNumber string `json:"TRACK_NUMBER"`
	TrackToken  string `json:"TRACK_TOKEN"`
}

func (s *Song) GetTitle() string {
	songTitle := s.Title
	if s.Version != "" {
		songTitle = fmt.Sprintf("%s %s", s.Title, s.Version)
	}

	return songTitle
}

func (s *Song) GetFileName(resourceType string, song *Song, media *Media) string {
	ext := "mp3"
	if media.Data[0].Media[0].Format == "FLAC" {
		ext = "flac"
	}
	trackNumber := ""
	if resourceType == "album" {
		trackNumber = song.TrackNumber + "."
	}

	fileName := fmt.Sprintf("%s %s - %s.%s", trackNumber, s.Artist, s.GetTitle(), ext)
	fileName, _ = filenamify.Filenamify(fileName, filenamify.Options{})

	return fileName
}
