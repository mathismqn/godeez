package deezer

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/flytam/filenamify"
)

type Contributors struct {
	MainArtists []string `json:"main_artist"`
	Composers   []string `json:"composer"`
	Authors     []string `json:"author"`
}

func (c *Contributors) UnmarshalJSON(data []byte) error {
	if string(data) == "[]" {
		*c = Contributors{}
		return nil
	}

	type Alias Contributors
	aux := (*Alias)(c)

	return json.Unmarshal(data, aux)
}

type Song struct {
	ID                  string       `json:"SNG_ID"`
	Artist              string       `json:"ART_NAME"`
	Title               string       `json:"SNG_TITLE"`
	Version             string       `json:"VERSION"`
	Cover               string       `json:"ALB_PICTURE"`
	AlbumTitle          string       `json:"ALB_TITLE"`
	Contributors        Contributors `json:"SNG_CONTRIBUTORS"`
	Duration            string       `json:"DURATION"`
	Gain                string       `json:"GAIN"`
	ISRC                string       `json:"ISRC"`
	TrackNumber         string       `json:"TRACK_NUMBER"`
	TrackToken          string       `json:"TRACK_TOKEN"`
	DiskNumber          string       `json:"DISK_NUMBER"`
	Copyright           string       `json:"COPYRIGHT"`
	PhysicalReleaseDate string       `json:"PHYSICAL_RELEASE_DATE"`
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
		trackNumber = song.TrackNumber + ". "
	}

	fileName := fmt.Sprintf("%s%s - %s.%s", trackNumber, s.Artist, s.GetTitle(), ext)
	fileName, _ = filenamify.Filenamify(fileName, filenamify.Options{})

	return fileName
}

// GetOrganizedPath returns the tree-structured path for this song: Artist/Album/Song
func (s *Song) GetOrganizedPath(baseOutputDir string, media *Media) string {
	ext := "mp3"
	if len(media.Data) > 0 && len(media.Data[0].Media) > 0 && media.Data[0].Media[0].Format == "FLAC" {
		ext = "flac"
	}

	// Use album artist if available, fallback to song artist
	artistName := s.Artist

	// Use album title if available, fallback to "Unknown Album"
	albumName := s.AlbumTitle
	if albumName == "" {
		albumName = "Unknown Album"
	}

	// Create filename with track number for albums
	trackNumber := ""
	if s.TrackNumber != "" {
		trackNumber = s.TrackNumber + ". "
	}
	fileName := fmt.Sprintf("%s%s.%s", trackNumber, s.GetTitle(), ext)

	// Sanitize all path components
	artistName, _ = filenamify.Filenamify(artistName, filenamify.Options{})
	albumName, _ = filenamify.Filenamify(albumName, filenamify.Options{})
	fileName, _ = filenamify.Filenamify(fileName, filenamify.Options{})

	return path.Join(baseOutputDir, artistName, albumName, fileName)
}
