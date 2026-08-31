package deezer

import (
	"encoding/json"
	"fmt"
	"strconv"
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

// Number is a numeric gateway field, which Deezer is inconsistent about
// quoting. A plain string field fails the whole response on the one entry
// that sends a bare number, so both forms are accepted as text. Every numeric
// field of a gateway struct is one of these, not only the ones already seen
// unquoted.
//
// An unquoted value has to be a JSON number. Its literal text is kept, which
// preserves a personal upload's negative id; an object or a boolean fails the
// decode rather than travelling on to be a filename or a Blowfish key.
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

	// The value is already known to be valid JSON, so its first byte is
	// enough to tell a number from the kinds that are refused.
	if c := data[0]; c != '-' && (c < '0' || c > '9') {
		return fmt.Errorf("cannot unmarshal %s into a gateway number", data)
	}

	*n = Number(data)

	return nil
}

// Int returns the number as an int, and whether it is one. Deezer omits these
// fields often enough that callers treat a false as unknown rather than as a
// failure.
func (n Number) Int() (int, bool) {
	i, err := strconv.Atoi(string(n))
	if err != nil {
		return 0, false
	}

	return i, true
}

type Track struct {
	ID           Number       `json:"SNG_ID"`
	Artist       string       `json:"ART_NAME"`
	Title        string       `json:"SNG_TITLE"`
	Version      string       `json:"VERSION"`
	Cover        string       `json:"ALB_PICTURE"`
	Contributors Contributors `json:"SNG_CONTRIBUTORS"`
	Duration     Number       `json:"DURATION"`
	Gain         Number       `json:"GAIN"`
	ISRC         string       `json:"ISRC"`
	TrackNumber  Number       `json:"TRACK_NUMBER"`
	DiscNumber   Number       `json:"DISK_NUMBER"`
	TrackToken   string       `json:"TRACK_TOKEN"`

	// Type tells a catalogue track apart from a personal upload; see
	// IsPersonalUpload.
	Type Number `json:"TYPE"`

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

// IsPersonalUpload reports whether this entry is a file the playlist owner
// uploaded rather than a track from Deezer's catalogue. Deezer marks those
// with TYPE 1 and serves them with no streaming rights at all.
//
// An upload still carries a TRACK_TOKEN like any other entry, so the token
// cannot be used to tell the two apart.
func (t *Track) IsPersonalUpload() bool {
	return t.Type == "1"
}
