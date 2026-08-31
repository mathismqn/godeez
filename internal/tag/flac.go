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
		t.addTag("TRACKTOTAL", m.TrackTotal)
		t.addTag("DISCNUMBER", m.DiscNumber)
		t.addTag("DISCTOTAL", m.DiscTotal)
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
	// index 0 means no comment block was found: a valid flac always starts
	// with STREAMINFO, so a real Vorbis comment can never be the first block.
	// Anything else is the index of the block being replaced.
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

// addTag appends a Vorbis comment. Vorbis allows repeated keys, so this adds
// to whatever the file already had rather than replacing it; re-tagging a
// file that was already tagged would therefore duplicate entries. That does
// not arise in practice because godeez only tags files it just downloaded.
//
// The key is written twice for the musical key: KEY is the common spelling
// and INITIALKEY is what several DJ applications look for.
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
