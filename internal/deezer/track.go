package deezer

import (
	"encoding/json"
	"fmt"
	"strconv"

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

type Track struct {
	ID           string       `json:"SNG_ID"`
	Artist       string       `json:"ART_NAME"`
	Title        string       `json:"SNG_TITLE"`
	Version      string       `json:"VERSION"`
	Cover        string       `json:"ALB_PICTURE"`
	Contributors Contributors `json:"SNG_CONTRIBUTORS"`
	Duration     string       `json:"DURATION"`
	Gain         string       `json:"GAIN"`
	ISRC         string       `json:"ISRC"`
	TrackNumber  string       `json:"TRACK_NUMBER"`
	TrackToken   string       `json:"TRACK_TOKEN"`
}

func (t *Track) FullTitle() string {
	if t.Version != "" {
		return t.Title + " " + t.Version
	}
	return t.Title
}

func (t *Track) Filename(kind Kind, mediaFormat string) string {
	ext := "mp3"
	if mediaFormat == "FLAC" {
		ext = "flac"
	}

	prefix := ""
	if kind == KindAlbum {
		if n, err := strconv.Atoi(t.TrackNumber); err == nil {
			prefix = fmt.Sprintf("%02d. ", n)
		} else {
			prefix = t.TrackNumber + ". "
		}
	}

	base := fmt.Sprintf("%s%s - %s", prefix, t.Artist, t.FullTitle())
	base, _ = filenamify.Filenamify(base, filenamify.Options{MaxLength: 255})
	base = truncateBytes(base, 255-len(ext)-1-len("-id3v2"))

	return base + "." + ext
}

func truncateBytes(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}

	last := 0
	for i := range s {
		if i > maxLen {
			return s[:last]
		}
		last = i
	}
	return s[:last]
}
