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

// UnmarshalJSON tolerates the empty array Deezer sends when a track has no
// contributors. The field is an object in every other case, so decoding it
// straight into the struct fails on those tracks.
func (c *Contributors) UnmarshalJSON(data []byte) error {
	if string(data) == "[]" {
		*c = Contributors{}
		return nil
	}

	// Alias drops the method set, so this Unmarshal does not recurse.
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

// Filename builds the on-disk name for the track, sanitised for the current
// filesystem. Album downloads get a zero padded track number prefix so the
// directory sorts in playing order; the other kinds have no meaningful
// ordering to preserve.
func (t *Track) Filename(kind Kind, format string) string {
	ext := "mp3"
	switch format {
	case "FLAC":
		ext = "flac"
	case "WAV":
		ext = "wav"
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
	// 255 bytes is the per-component limit on ext4 and APFS. The budget also
	// has to cover the extension, its dot, and the "-id3v2" suffix the tagging
	// library appends to its temporary file: without that headroom, tagging a
	// long title fails after the download has already succeeded.
	base = truncateBytes(base, 255-len(ext)-1-len("-id3v2"))

	return base + "." + ext
}

// truncateBytes shortens s to at most maxLen bytes without splitting a rune.
// The limit is in bytes because that is what filesystems enforce, but cutting
// mid-rune would leave an invalid UTF-8 name, so it backs up to the last rune
// boundary that fits.
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
