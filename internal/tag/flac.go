package tag

import (
	"os"
	"strings"

	"github.com/go-flac/flacpicture/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
)

type flacTagger struct {
	file  *flac.File
	cmts  *flacvorbis.MetaDataBlockVorbisComment
	index int
	path  string
}

func (t *flacTagger) write(m Metadata) error {
	if m.Album != nil {
		date := m.Album.ReleaseDate
		if parts := strings.Split(date, "-"); len(parts) == 3 {
			date = parts[0]
		}

		t.addTag("TRACKNUMBER", m.TrackNumber)
		t.addTag("ALBUMARTIST", m.Album.Artist)
		t.addTag("ALBUM", m.Album.Title)
		t.addTag("PUBLISHER", m.Album.Label)
		t.addTag("ORIGINALDATE", m.Album.OriginalReleaseDate)
		t.addTag("DATE", date)
		t.addTag("COMMENT", m.Album.ProducerLine)
		t.addTag("COPYRIGHT", m.Album.Copyright)
	}

	t.addTag("ARTIST", m.Artists)
	t.addTag("TITLE", m.Title)
	t.addTag("COMPOSER", m.Composers)
	t.addTag("LYRICIST", m.Lyricists)
	t.addTag("GENRE", m.Genre)
	t.addTag("REPLAYGAIN_TRACK_GAIN", m.Gain)
	t.addTag("ISRC", m.ISRC)
	t.addTag("BPM", m.BPM)
	t.addTag("KEY", m.Key)
	t.addTag("INITIALKEY", m.Key)

	cmtsMeta := t.cmts.Marshal()
	if t.index > 0 {
		t.file.Meta[t.index] = &cmtsMeta
	} else {
		t.file.Meta = append(t.file.Meta, &cmtsMeta)
	}

	if len(m.Cover) > 0 {
		if picture, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "Front cover", m.Cover, "image/jpeg"); err == nil {
			pictureMeta := picture.Marshal()
			t.file.Meta = append(t.file.Meta, &pictureMeta)
		}
	}

	tmpPath := t.path + ".tmp"
	if err := t.file.Save(tmpPath); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, t.path)
}

func (t *flacTagger) addTag(name, value string) {
	if value != "" {
		t.cmts.Add(name, value)
	}
}

func extractFLACComment(file *flac.File) (*flacvorbis.MetaDataBlockVorbisComment, int) {
	for idx, meta := range file.Meta {
		if meta.Type == flac.VorbisComment {
			cmt, err := flacvorbis.ParseFromMetaDataBlock(*meta)
			if err == nil {
				return cmt, idx
			}
		}
	}
	return nil, 0
}
