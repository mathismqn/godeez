package deezer

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFullTitle(t *testing.T) {
	tests := []struct {
		title   string
		version string
		want    string
	}{
		{"Song", "", "Song"},
		{"Song", "(Remix)", "Song (Remix)"},
	}

	for _, tt := range tests {
		track := &Track{Title: tt.title, Version: tt.version}
		if got := track.FullTitle(); got != tt.want {
			t.Errorf("FullTitle() = %q, want %q", got, tt.want)
		}
	}
}

func TestContributorsUnmarshalJSON(t *testing.T) {
	var c Contributors
	if err := json.Unmarshal([]byte("[]"), &c); err != nil {
		t.Fatalf("unmarshal empty array: %v", err)
	}
	if len(c.MainArtists) != 0 || len(c.Composers) != 0 || len(c.Authors) != 0 {
		t.Errorf("expected empty contributors, got %+v", c)
	}

	data := `{"main_artist":["A","B"],"composer":["C"],"author":["D"]}`
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		t.Fatalf("unmarshal object: %v", err)
	}
	if len(c.MainArtists) != 2 || c.MainArtists[0] != "A" || len(c.Composers) != 1 || len(c.Authors) != 1 {
		t.Errorf("unexpected contributors: %+v", c)
	}
}

func TestTrackUnmarshalFallback(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []Number
	}{
		{
			name: "no fallback",
			data: `{"SNG_ID":"2358247075","ISRC":"DE1FB2300002"}`,
			want: nil,
		},
		{
			name: "one fallback",
			data: `{"SNG_ID":"2358247065","FALLBACK":{"SNG_ID":"2134121047"}}`,
			want: []Number{"2134121047"},
		},
		{
			name: "nested fallback",
			data: `{"SNG_ID":"a","FALLBACK":{"SNG_ID":"b","FALLBACK":{"SNG_ID":"c"}}}`,
			want: []Number{"b", "c"},
		},
		{
			name: "empty fallback object",
			data: `{"SNG_ID":"a","FALLBACK":{}}`,
			want: []Number{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var track Track
			if err := json.Unmarshal([]byte(tt.data), &track); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			var got []Number
			for next := track.Fallback; next != nil; next = next.Fallback {
				got = append(got, next.ID)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fallback chain = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackUnmarshalFallbackToken(t *testing.T) {
	var track Track
	data := `{"SNG_ID":"2358247065","TRACK_TOKEN":"parent","FALLBACK":{"SNG_ID":"2134121047","TRACK_TOKEN":"duplicate"}}`
	if err := json.Unmarshal([]byte(data), &track); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if track.Fallback.TrackToken != "duplicate" {
		t.Errorf("Fallback.TrackToken = %q, want duplicate", track.Fallback.TrackToken)
	}
}

// TestTrackNumberDecoding guards every field the gateway is inconsistent
// about quoting: an unquoted one used to fail the whole response rather than
// the one field.
func TestTrackNumberDecoding(t *testing.T) {
	tests := []struct {
		name string
		json string
		want Track
	}{
		{
			name: "quoted",
			json: `{"SNG_ID":"2358247075","DURATION":"213","GAIN":"-11.2","TRACK_NUMBER":"7","DISK_NUMBER":"2"}`,
			want: Track{ID: "2358247075", Duration: "213", Gain: "-11.2", TrackNumber: "7", DiscNumber: "2"},
		},
		{
			name: "bare",
			json: `{"SNG_ID":2358247075,"DURATION":213,"GAIN":-11.2,"TRACK_NUMBER":7,"DISK_NUMBER":2}`,
			want: Track{ID: "2358247075", Duration: "213", Gain: "-11.2", TrackNumber: "7", DiscNumber: "2"},
		},
		{
			// A personal upload's id, which is what made a whole playlist
			// carrying one fail to decode.
			name: "bare negative id",
			json: `{"SNG_ID":-3002903542}`,
			want: Track{ID: "-3002903542"},
		},
		{
			name: "null",
			json: `{"SNG_ID":null,"DURATION":null,"GAIN":null,"TRACK_NUMBER":null,"DISK_NUMBER":null}`,
			want: Track{},
		},
		{
			name: "absent",
			json: `{}`,
			want: Track{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var track Track
			if err := json.Unmarshal([]byte(tt.json), &track); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(track, tt.want) {
				t.Errorf("Track = %+v, want %+v", track, tt.want)
			}
		})
	}
}

// TestTrackNumberRejectsNonNumbers keeps an unquoted object or boolean from
// being stored as literal text.
func TestTrackNumberRejectsNonNumbers(t *testing.T) {
	for _, data := range []string{`{"SNG_ID":{}}`, `{"SNG_ID":[]}`, `{"SNG_ID":true}`} {
		var track Track
		if err := json.Unmarshal([]byte(data), &track); err == nil {
			t.Errorf("Unmarshal(%s) = nil error, want a failure (ID = %q)", data, track.ID)
		}
	}
}
