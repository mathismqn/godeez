package download

import (
	"testing"

	"github.com/mathismqn/godeez/internal/deezer"
)

// TestBuildTagMetadataArtistFallback checks the artist tag goes through
// ArtistNames, whose cases are covered in the deezer package.
func TestBuildTagMetadataArtistFallback(t *testing.T) {
	got := buildTagMetadata(&deezer.Playlist{}, &deezer.Track{Artist: "Lead"}, nil, bpmKey{}, "", discLayout{})
	if got.Artists != "Lead" {
		t.Errorf("Artists = %q, want %q", got.Artists, "Lead")
	}
}
