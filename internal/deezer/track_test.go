package deezer

import (
	"encoding/json"
	"reflect"
	"strings"
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

func TestFilename(t *testing.T) {
	tests := []struct {
		name        string
		track       Track
		kind        Kind
		mediaFormat string
		want        string
	}{
		{
			name:        "album with numeric track number",
			track:       Track{Artist: "Artist", Title: "Song", TrackNumber: "1"},
			kind:        KindAlbum,
			mediaFormat: "MP3_320",
			want:        "01. Artist - Song.mp3",
		},
		{
			name:        "album with non-numeric track number",
			track:       Track{Artist: "Artist", Title: "Song", TrackNumber: "A"},
			kind:        KindAlbum,
			mediaFormat: "MP3_320",
			want:        "A. Artist - Song.mp3",
		},
		{
			name:        "playlist has no prefix",
			track:       Track{Artist: "Artist", Title: "Song", TrackNumber: "1"},
			kind:        KindPlaylist,
			mediaFormat: "MP3_128",
			want:        "Artist - Song.mp3",
		},
		{
			name:        "flac extension",
			track:       Track{Artist: "Artist", Title: "Song"},
			kind:        KindTrack,
			mediaFormat: "FLAC",
			want:        "Artist - Song.flac",
		},
		{
			name:        "wav extension",
			track:       Track{Artist: "Artist", Title: "Song"},
			kind:        KindTrack,
			mediaFormat: "WAV",
			want:        "Artist - Song.wav",
		},
		{
			name:        "version appended",
			track:       Track{Artist: "Artist", Title: "Song", Version: "(Live)"},
			kind:        KindTrack,
			mediaFormat: "MP3_320",
			want:        "Artist - Song (Live).mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.track.Filename(tt.kind, tt.mediaFormat); got != tt.want {
				t.Errorf("Filename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFilenameSanitizesSeparators(t *testing.T) {
	track := Track{Artist: "AC/DC", Title: "Song"}
	got := track.Filename(KindTrack, "MP3_320")

	if strings.ContainsRune(got, '/') {
		t.Errorf("Filename() contains a path separator: %q", got)
	}
	if !strings.HasSuffix(got, ".mp3") {
		t.Errorf("Filename() = %q, want .mp3 suffix", got)
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
