package download

import (
	"strconv"
	"strings"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/tag"
)

// buildTagMetadata assembles the tag set for a finished file.
//
// Album tags are written for every kind of download: a track from a
// playlist still belongs to a release, and its entry names it. Only the
// totals, derived from the whole release, need an album download.
func buildTagMetadata(resource deezer.Resource, track *deezer.Track, cover []byte, bpm bpmKey, genre string, discs discLayout) tag.Metadata {
	album, _ := resource.(*deezer.Album)

	m := tag.Metadata{
		Title:       track.FullTitle(),
		Artists:     strings.Join(track.ArtistNames(), ", "),
		Composers:   strings.Join(track.Contributors.Composers, ", "),
		Lyricists:   strings.Join(track.Contributors.Authors, ", "),
		Genre:       genre,
		BPM:         bpm.BPM,
		Key:         bpm.Key,
		TrackNumber: string(track.TrackNumber),
		Duration:    string(track.Duration),
		Gain:        string(track.Gain),
		ISRC:        track.ISRC,
		Cover:       cover,
		Album:       albumMetadata(album, track),
	}

	if album != nil {
		// Deezer's track number is a position within its own disc, so its
		// total is the size of that disc rather than of the release. A single
		// disc release has no disc position worth recording.
		disc := discNumber(track)
		m.TrackTotal = strconv.Itoa(discs.trackTotal(disc))
		if discs.multiDisc() {
			m.DiscNumber = disc
			m.DiscTotal = strconv.Itoa(discs.discTotal())
		}
	}

	return m
}

// albumMetadata returns the release track belongs to, from the album page
// when there is one and from the track entry otherwise, or nil when Deezer
// names no album.
//
// The track entry does not carry the album artist, and the track's own
// artist is no substitute: on a compilation it would replace "Various
// Artists" and players grouping by album artist would split the release.
func albumMetadata(album *deezer.Album, track *deezer.Track) *tag.AlbumMetadata {
	if album != nil {
		data := album.Results.Data

		return &tag.AlbumMetadata{
			Artist:              data.Artist,
			Title:               data.Title,
			Label:               data.Label,
			OriginalReleaseDate: data.OriginalReleaseDate,
			ReleaseDate:         data.PhysicalReleaseDate,
			ProducerLine:        data.ProducerLine,
			Copyright:           data.Copyright,
		}
	}

	if track.AlbumTitle == "" {
		return nil
	}

	return &tag.AlbumMetadata{
		Title:       track.AlbumTitle,
		ReleaseDate: track.PhysicalReleaseDate,
		Copyright:   track.Copyright,
	}
}
