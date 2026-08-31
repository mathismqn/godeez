package download

import (
	"strconv"
	"strings"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/tag"
)

func buildTagMetadata(resource deezer.Resource, track *deezer.Track, cover []byte, bpm bpmKey, genre string, discs discLayout) tag.Metadata {
	m := tag.Metadata{
		Title:       track.FullTitle(),
		Artists:     strings.Join(track.Contributors.MainArtists, ", "),
		Composers:   strings.Join(track.Contributors.Composers, ", "),
		Lyricists:   strings.Join(track.Contributors.Authors, ", "),
		Genre:       genre,
		BPM:         bpm.BPM,
		Key:         bpm.Key,
		TrackNumber: string(track.TrackNumber),
		Duration:    track.Duration,
		Gain:        track.Gain,
		ISRC:        track.ISRC,
		Cover:       cover,
	}

	if album, ok := resource.(*deezer.Album); ok {
		data := album.Results.Data
		m.Album = &tag.AlbumMetadata{
			Artist:              data.Artist,
			Title:               data.Title,
			Label:               data.Label,
			OriginalReleaseDate: data.OriginalReleaseDate,
			ReleaseDate:         data.PhysicalReleaseDate,
			ProducerLine:        data.ProducerLine,
			Copyright:           data.Copyright,
		}

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
