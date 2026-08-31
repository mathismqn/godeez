package download

import (
	"strings"
	"testing"

	"github.com/mathismqn/godeez/internal/deezer"
)

func TestTrackFilename(t *testing.T) {
	twoDiscs := discLayoutOf("1", "2")

	tests := []struct {
		name        string
		track       deezer.Track
		kind        deezer.Kind
		mediaFormat string
		discs       discLayout
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
		{
			name:        "single disc album has no disc prefix",
			track:       deezer.Track{Artist: "Artist", Title: "Song", TrackNumber: "1", DiscNumber: "1"},
			kind:        deezer.KindAlbum,
			mediaFormat: "MP3_320",
			discs:       discLayoutOf("1"),
			want:        "01. Artist - Song.mp3",
		},
		{
			name:        "multi disc album prefixes the first disc",
			track:       deezer.Track{Artist: "Artist", Title: "Song", TrackNumber: "1", DiscNumber: "1"},
			kind:        deezer.KindAlbum,
			mediaFormat: "MP3_320",
			discs:       twoDiscs,
			want:        "1-01. Artist - Song.mp3",
		},
		{
			name:        "multi disc album keeps the second disc apart",
			track:       deezer.Track{Artist: "Artist", Title: "Song", TrackNumber: "1", DiscNumber: "2"},
			kind:        deezer.KindAlbum,
			mediaFormat: "MP3_320",
			discs:       twoDiscs,
			want:        "2-01. Artist - Song.mp3",
		},
		{
			name:        "multi disc album with non-numeric disc number",
			track:       deezer.Track{Artist: "Artist", Title: "Song", TrackNumber: "1", DiscNumber: "B"},
			kind:        deezer.KindAlbum,
			mediaFormat: "MP3_320",
			discs:       discLayoutOf("A", "B"),
			want:        "B-01. Artist - Song.mp3",
		},
		{
			name:        "multi disc playlist still has no prefix",
			track:       deezer.Track{Artist: "Artist", Title: "Song", TrackNumber: "1", DiscNumber: "2"},
			kind:        deezer.KindPlaylist,
			mediaFormat: "MP3_320",
			discs:       twoDiscs,
			want:        "Artist - Song.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trackFilename(&tt.track, tt.kind, tt.mediaFormat, tt.discs); got != tt.want {
				t.Errorf("trackFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrackFilenameSanitizesSeparators(t *testing.T) {
	track := deezer.Track{Artist: "AC/DC", Title: "Song"}
	got := trackFilename(&track, deezer.KindTrack, "MP3_320", nil)

	if strings.ContainsRune(got, '/') {
		t.Errorf("trackFilename() contains a path separator: %q", got)
	}
	if !strings.HasSuffix(got, ".mp3") {
		t.Errorf("trackFilename() = %q, want .mp3 suffix", got)
	}
}
