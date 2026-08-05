package tag

import (
	"fmt"
	"strconv"

	"github.com/bogem/id3v2/v2"
)

type id3v2Tagger struct {
	tag *id3v2.Tag
}

func (t *id3v2Tagger) write(m Metadata) error {
	defer t.tag.Close()

	duration, err := strconv.Atoi(m.Duration)
	if err != nil {
		return err
	}
	length := fmt.Sprintf("%d", duration*1000)

	if m.Album != nil {
		t.addTag("TRCK", m.TrackNumber)
		t.addTag("TPE2", m.Album.Artist)
		t.addTag("TALB", m.Album.Title)
		t.addTag("TPUB", m.Album.Label)
		t.addTag("TDOR", m.Album.OriginalReleaseDate)
		t.addTag("TYER", m.Album.ReleaseDate)
		t.addTag("COMM", m.Album.ProducerLine)
		t.addTag("TCOP", m.Album.Copyright)
	}

	t.addTag("TPE1", m.Artists)
	t.addTag("TIT2", m.Title)
	t.addTag("TCOM", m.Composers)
	t.addTag("TEXT", m.Lyricists)
	t.addTag("TCON", m.Genre)
	t.addTag("TLEN", length)
	t.addTag("TBPM", m.BPM)
	t.addTag("TKEY", m.Key)
	t.addTXXX("GAIN", m.Gain)
	t.addTXXX("ISRC", m.ISRC)

	t.tag.AddAttachedPicture(id3v2.PictureFrame{
		Encoding:    t.tag.DefaultEncoding(),
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Cover",
		Picture:     m.Cover,
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
