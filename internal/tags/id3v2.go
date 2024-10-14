package tags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/mathismqn/godeez/internal/deezer"
)

type ID3v2Tagger struct {
	Tag *id3v2.Tag
}

func (t *ID3v2Tagger) AddTags(resource deezer.Resource, song *deezer.Song, cover []byte, path string) error {
	defer t.Tag.Close()

	duration, _ := strconv.Atoi(song.Duration)
	song.Duration = fmt.Sprintf("%d", duration*1000)

	if album, ok := resource.(*deezer.Album); ok {
		t.addTag("TALB", album.Data.Title)
		t.addTag("TPE2", album.Data.Artist)
		t.addTag("TPUB", album.Data.Label)
		t.addTag("TDOR", album.Data.OriginalReleaseDate)
		t.addTag("TYER", album.Data.PhysicalReleaseDate)
		t.addTag("COMM", album.Data.ProducerLine)
		t.addTag("TRCK", song.TrackNumber)
	}

	t.addTag("TIT2", song.Title)
	t.addTag("TPE1", strings.Join(song.Contributors.MainArtists, ", "))
	t.addTag("TCOM", strings.Join(song.Contributors.Composers, ", "))
	t.addTag("TEXT", strings.Join(song.Contributors.Authors, ", "))
	t.addTag("TLEN", song.Duration)
	t.addTXXXTag("GAIN", song.Gain)
	t.addTXXXTag("ISRC", song.ISRC)

	frame := id3v2.PictureFrame{
		Encoding:    t.Tag.DefaultEncoding(),
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Cover",
		Picture:     cover,
	}
	t.Tag.AddAttachedPicture(frame)

	return t.Tag.Save()
}

func (t *ID3v2Tagger) addTag(name, value string) {
	if value != "" {
		t.Tag.AddTextFrame(name, t.Tag.DefaultEncoding(), value)
	}
}

func (t *ID3v2Tagger) addTXXXTag(description, value string) {
	if value != "" {
		udf := id3v2.UserDefinedTextFrame{
			Encoding:    t.Tag.DefaultEncoding(),
			Description: description,
			Value:       value,
		}
		t.Tag.AddUserDefinedTextFrame(udf)
	}
}
