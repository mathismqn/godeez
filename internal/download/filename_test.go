package download

import (
	"strings"
	"testing"

	"github.com/mathismqn/godeez/internal/deezer"
)

func TestTrackFilename(t *testing.T) {
	tests := []struct {
		name        string
		track       deezer.Track
		kind        deezer.Kind
		mediaFormat string
		want        string
	}{
		{
			name:        "album with numeric track number",
			track:       deezer.Track{Artist: "Artist", Title: "Song", TrackNumber: "1"},
			kind:        deezer.KindAlbum,
			mediaFormat: "MP3_320",
			want:        "01. Artist - Song.mp3",
		},
		{
			name:        "album with non-numeric track number",
			track:       deezer.Track{Artist: "Artist", Title: "Song", TrackNumber: "A"},
			kind:        deezer.KindAlbum,
			mediaFormat: "MP3_320",
			want:        "A. Artist - Song.mp3",
		},
		{
			name:        "playlist has no prefix",
			track:       deezer.Track{Artist: "Artist", Title: "Song", TrackNumber: "1"},
			kind:        deezer.KindPlaylist,
			mediaFormat: "MP3_128",
			want:        "Artist - Song.mp3",
		},
		{
			name:        "flac extension",
			track:       deezer.Track{Artist: "Artist", Title: "Song"},
			kind:        deezer.KindTrack,
			mediaFormat: "FLAC",
			want:        "Artist - Song.flac",
		},
		{
			name:        "wav extension",
			track:       deezer.Track{Artist: "Artist", Title: "Song"},
			kind:        deezer.KindTrack,
			mediaFormat: "WAV",
			want:        "Artist - Song.wav",
		},
		{
			name:        "version appended",
			track:       deezer.Track{Artist: "Artist", Title: "Song", Version: "(Live)"},
			kind:        deezer.KindTrack,
			mediaFormat: "MP3_320",
			want:        "Artist - Song (Live).mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trackFilename(&tt.track, tt.kind, tt.mediaFormat); got != tt.want {
				t.Errorf("trackFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrackFilenameSanitizesSeparators(t *testing.T) {
	track := deezer.Track{Artist: "AC/DC", Title: "Song"}
	got := trackFilename(&track, deezer.KindTrack, "MP3_320")

	if strings.ContainsRune(got, '/') {
		t.Errorf("trackFilename() contains a path separator: %q", got)
	}
	if !strings.HasSuffix(got, ".mp3") {
		t.Errorf("trackFilename() = %q, want .mp3 suffix", got)
	}
}
