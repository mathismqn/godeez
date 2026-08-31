package tag

import (
	"testing"

	"github.com/bogem/id3v2/v2"
)

func TestJoinTotal(t *testing.T) {
	tests := []struct {
		name  string
		n     string
		total string
		want  string
	}{
		{name: "position and total", n: "3", total: "12", want: "3/12"},
		{name: "no total", n: "3", total: "", want: "3"},
		{name: "no position omits the frame", n: "", total: "12", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinTotal(tt.n, tt.total); got != tt.want {
				t.Errorf("joinTotal(%q, %q) = %q, want %q", tt.n, tt.total, got, tt.want)
			}
		})
	}
}

func TestApplyID3FramesPositions(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		wantTRCK string
		wantTPOS string
	}{
		{
			name: "multi disc album carries both positions",
			metadata: Metadata{
				TrackNumber: "1",
				TrackTotal:  "11",
				DiscNumber:  "1",
				DiscTotal:   "2",
				Album:       &AlbumMetadata{Title: "Album"},
			},
			wantTRCK: "1/11",
			wantTPOS: "1/2",
		},
		{
			name: "single disc album has a track total and no disc",
			metadata: Metadata{
				TrackNumber: "1",
				TrackTotal:  "12",
				Album:       &AlbumMetadata{Title: "Album"},
			},
			wantTRCK: "1/12",
			wantTPOS: "",
		},
		{
			name:     "a track outside an album has neither",
			metadata: Metadata{TrackNumber: "1", TrackTotal: "12"},
			wantTRCK: "",
			wantTPOS: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := id3v2.NewEmptyTag()
			applyID3Frames(tag, tt.metadata)

			if got := tag.GetTextFrame("TRCK").Text; got != tt.wantTRCK {
				t.Errorf("TRCK = %q, want %q", got, tt.wantTRCK)
			}
			if got := tag.GetTextFrame("TPOS").Text; got != tt.wantTPOS {
				t.Errorf("TPOS = %q, want %q", got, tt.wantTPOS)
			}
		})
	}
}
