package tags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/mathismqn/godeez/internal/deezer"
)

type id3v2Tagger struct {
	tag *id3v2.Tag
}

func (t *id3v2Tagger) addTags(resource deezer.Resource, track *deezer.Track, cover []byte, path, tempo, key, genre string) error {
	defer t.tag.Close()

	duration, err := strconv.Atoi(track.Duration)
	if err != nil {
		return err
	}
	track.Duration = fmt.Sprintf("%d", duration*1000)

	if album, ok := resource.(*deezer.Album); ok {
		t.addTag("TRCK", track.TrackNumber)
		t.addTag("TPE2", album.Results.Data.Artist)
		t.addTag("TALB", album.Results.Data.Title)
		t.addTag("TPUB", album.Results.Data.Label)
		t.addTag("TDOR", album.Results.Data.OriginalReleaseDate)
		t.addTag("TYER", album.Results.Data.PhysicalReleaseDate)
		t.addTag("COMM", album.Results.Data.ProducerLine)
		t.addTag("TCOP", album.Results.Data.Copyright)
	}

	t.addTag("TPE1", strings.Join(track.Contributors.MainArtists, ", "))
	t.addTag("TIT2", track.GetTitle())
	t.addTag("TCOM", strings.Join(track.Contributors.Composers, ", "))
	t.addTag("TEXT", strings.Join(track.Contributors.Authors, ", "))
	t.addTag("TCON", genre)
	t.addTag("TLEN", track.Duration)
	t.addTag("TBPM", tempo)
	t.addTag("TKEY", key)
	t.addTXXX("GAIN", track.Gain)
	t.addTXXX("ISRC", track.ISRC)

	t.tag.AddAttachedPicture(id3v2.PictureFrame{
		Encoding:    t.tag.DefaultEncoding(),
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Cover",
		Picture:     cover,
	})

	return t.tag.Save()
}

func (t *id3v2Tagger) addTag(name, value string) {
	if value != "" {
		t.tag.AddTextFrame(name, t.tag.DefaultEncoding(), value)
	}
}

func (t *id3v2Tagger) addTXXX(description, value string) {
	if value != "" {
		t.tag.AddUserDefinedTextFrame(id3v2.UserDefinedTextFrame{
			Encoding:    t.tag.DefaultEncoding(),
			Description: description,
			Value:       value,
		})
	}
}
