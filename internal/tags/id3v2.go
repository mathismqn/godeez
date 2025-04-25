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

func (t *ID3v2Tagger) AddTags(resource deezer.Resource, song *deezer.Song, cover []byte, path, tempo, key string) error {
	defer t.Tag.Close()

	duration, _ := strconv.Atoi(song.Duration)
	song.Duration = fmt.Sprintf("%d", duration*1000)

	if album, ok := resource.(*deezer.Album); ok {
		t.addTag("TALB", album.Results.Data.Title)
		t.addTag("TPE2", album.Results.Data.Artist)
		t.addTag("TPUB", album.Results.Data.Label)
		t.addTag("TDOR", album.Results.Data.OriginalReleaseDate)
		t.addTag("TYER", album.Results.Data.PhysicalReleaseDate)
		t.addTag("COMM", album.Results.Data.ProducerLine)
		t.addTag("TRCK", song.TrackNumber)
	}

	t.addTag("TIT2", song.Title)
	t.addTag("TPE1", strings.Join(song.Contributors.MainArtists, ", "))
	t.addTag("TCOM", strings.Join(song.Contributors.Composers, ", "))
	t.addTag("TEXT", strings.Join(song.Contributors.Authors, ", "))
	t.addTag("TLEN", song.Duration)
	t.addTXXXTag("GAIN", song.Gain)
	t.addTXXXTag("ISRC", song.ISRC)

	t.addTag("TBPM", tempo)
	t.addTag("TKEY", key)

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
