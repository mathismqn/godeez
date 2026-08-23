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
		want []string
	}{
		{
			name: "no fallback",
			data: `{"SNG_ID":"2358247075","ISRC":"DE1FB2300002"}`,
			want: nil,
		},
		{
			name: "one fallback",
			data: `{"SNG_ID":"2358247065","FALLBACK":{"SNG_ID":"2134121047"}}`,
			want: []string{"2134121047"},
		},
		{
			name: "nested fallback",
			data: `{"SNG_ID":"a","FALLBACK":{"SNG_ID":"b","FALLBACK":{"SNG_ID":"c"}}}`,
			want: []string{"b", "c"},
		},
		{
			name: "empty fallback object",
			data: `{"SNG_ID":"a","FALLBACK":{}}`,
			want: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var track Track
			if err := json.Unmarshal([]byte(tt.data), &track); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			var got []string
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
