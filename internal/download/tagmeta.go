package download

import (
	"strings"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/tag"
)

func buildTagMetadata(resource deezer.Resource, track *deezer.Track, cover []byte, bpm bpmKey, genre string) tag.Metadata {
	m := tag.Metadata{
		Title:       track.GetTitle(),
		Artists:     strings.Join(track.Contributors.MainArtists, ", "),
		Composers:   strings.Join(track.Contributors.Composers, ", "),
		Lyricists:   strings.Join(track.Contributors.Authors, ", "),
		Genre:       genre,
		BPM:         bpm.BPM,
		Key:         bpm.Key,
		TrackNumber: track.TrackNumber,
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
	}

	return m
}
