package download

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/tag"
)

// TestBuildTagMetadataArtistFallback checks the artist tag goes through
// ArtistNames, whose cases are covered in the deezer package.
func TestBuildTagMetadataArtistFallback(t *testing.T) {
	got := buildTagMetadata(&deezer.Playlist{}, &deezer.Track{Artist: "Lead"}, nil, bpmKey{}, "", discLayout{})
	if got.Artists != "Lead" {
		t.Errorf("Artists = %q, want %q", got.Artists, "Lead")
	}
}

// TestBuildTagMetadataAlbumOfTrack pins the album tags of a download whose
// resource is not an album: what the track entry carries, with no album
// artist and no totals.
func TestBuildTagMetadataAlbumOfTrack(t *testing.T) {
	track := &deezer.Track{
		Artist:              "The Artist",
		AlbumTitle:          "The Album",
		PhysicalReleaseDate: "2020-05-01",
		Copyright:           "(C) 2020 The Label",
		TrackNumber:         "4",
	}

	got := buildTagMetadata(&deezer.Playlist{}, track, nil, bpmKey{}, "", discLayout{})
	want := tag.AlbumMetadata{Title: "The Album", ReleaseDate: "2020-05-01", Copyright: "(C) 2020 The Label"}
	if got.Album == nil || !reflect.DeepEqual(*got.Album, want) {
		t.Errorf("Album = %+v, want %+v", got.Album, want)
	}
	if got.TrackNumber != "4" || got.TrackTotal != "" {
		t.Errorf("track = %q/%q, want %q with no total", got.TrackNumber, got.TrackTotal, "4")
	}
}

// TestBuildTagMetadataWithoutAlbum keeps a track Deezer sends no album for
// from being tagged with a blank ALBUM frame.
func TestBuildTagMetadataWithoutAlbum(t *testing.T) {
	got := buildTagMetadata(&deezer.Playlist{}, &deezer.Track{Artist: "The Artist"}, nil, bpmKey{}, "", discLayout{})
	if got.Album != nil {
		t.Errorf("Album = %+v, want nil", got.Album)
	}
}

// TestBuildTagMetadataAlbumResource checks the album page wins over the
// track entry, album artist included, and still supplies the disc layout.
func TestBuildTagMetadataAlbumResource(t *testing.T) {
	album := &deezer.Album{}
	data := `{"results":{"DATA":{"ALB_TITLE":"The Album","ART_NAME":"Various Artists","LABEL_NAME":"The Label","ORIGINAL_RELEASE_DATE":"1999-01-01","PHYSICAL_RELEASE_DATE":"2020-05-01","PRODUCER_LINE":"Produced by Someone","COPYRIGHT":"(C) 2020 The Label"}}}`
	if err := json.Unmarshal([]byte(data), album); err != nil {
		t.Fatalf("unmarshal album: %v", err)
	}

	track := &deezer.Track{Artist: "Track Artist", AlbumTitle: "Other", DiscNumber: "2"}
	discs := discLayoutOf("1", "1", "2", "2", "2")

	got := buildTagMetadata(album, track, nil, bpmKey{}, "", discs)
	want := tag.AlbumMetadata{
		Artist:              "Various Artists",
		Title:               "The Album",
		Label:               "The Label",
		OriginalReleaseDate: "1999-01-01",
		ReleaseDate:         "2020-05-01",
		ProducerLine:        "Produced by Someone",
		Copyright:           "(C) 2020 The Label",
	}
	if got.Album == nil || !reflect.DeepEqual(*got.Album, want) {
		t.Errorf("Album = %+v, want %+v", got.Album, want)
	}
	// The total is the size of the track's own disc, not of the release.
	if got.TrackTotal != "3" {
		t.Errorf("TrackTotal = %q, want %q", got.TrackTotal, "3")
	}
	if got.DiscNumber != "2" || got.DiscTotal != "2" {
		t.Errorf("disc = %q/%q, want %q/%q", got.DiscNumber, got.DiscTotal, "2", "2")
	}
}
