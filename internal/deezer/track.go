package deezer

import (
	"encoding/json"
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

// Number is a position from the gateway, which is inconsistent about whether
// it quotes them. A plain string field fails the whole album on the one
// response that sends a bare number, so both forms are accepted as text.
type Number string

func (n *Number) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*n = Number(s)

		return nil
	}

	if string(data) == "null" {
		*n = ""
		return nil
	}

	*n = Number(data)

	return nil
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
	TrackNumber  Number       `json:"TRACK_NUMBER"`
	DiscNumber   Number       `json:"DISK_NUMBER"`
	TrackToken   string       `json:"TRACK_TOKEN"`

	// Fallback is the readable duplicate Deezer points at when this entry
	// has no streaming rights of its own.
	Fallback *Track `json:"FALLBACK"`
}

func (t *Track) FullTitle() string {
	if t.Version != "" {
		return t.Title + " " + t.Version
	}
	return t.Title
}
