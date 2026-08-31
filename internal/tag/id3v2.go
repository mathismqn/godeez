package tag

import (
	"strconv"
	"strings"

	"github.com/bogem/id3v2/v2"
)

type id3v2Tagger struct {
	tag *id3v2.Tag
}

func (t *id3v2Tagger) write(m Metadata) error {
	defer t.tag.Close()

	applyID3Frames(t.tag, m)

	return t.tag.Save()
}

func applyID3Frames(tag *id3v2.Tag, m Metadata) {
	if m.Album != nil {
		year := m.Album.ReleaseDate
		if parts := strings.Split(year, "-"); len(parts) == 3 {
			year = parts[0]
		}

		addID3Text(tag, "TRCK", joinTotal(m.TrackNumber, m.TrackTotal))
		addID3Text(tag, "TPOS", joinTotal(m.DiscNumber, m.DiscTotal))
		addID3Text(tag, "TPE2", m.Album.Artist)
		addID3Text(tag, "TALB", m.Album.Title)
		addID3Text(tag, "TPUB", m.Album.Label)
		addID3Text(tag, "TDOR", m.Album.OriginalReleaseDate)
		addID3Text(tag, "TYER", year)
		addID3Comment(tag, m.Album.ProducerLine)
		addID3Text(tag, "TCOP", m.Album.Copyright)
	}

	addID3Text(tag, "TPE1", m.Artists)
	addID3Text(tag, "TIT2", m.Title)
	addID3Text(tag, "TCOM", m.Composers)
	addID3Text(tag, "TEXT", m.Lyricists)
	addID3Text(tag, "TCON", m.Genre)
	if duration, err := strconv.Atoi(m.Duration); err == nil {
		addID3Text(tag, "TLEN", strconv.Itoa(duration*1000))
	}
	addID3Text(tag, "TBPM", m.BPM)
	addID3Text(tag, "TKEY", m.Key)
	addID3TXXX(tag, "GAIN", m.Gain)
	addID3TXXX(tag, "ISRC", m.ISRC)

	if len(m.Cover) > 0 {
		tag.AddAttachedPicture(id3v2.PictureFrame{
			Encoding:    tag.DefaultEncoding(),
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Description: "Cover",
			Picture:     m.Cover,
		})
	}
}

// joinTotal renders a position the way ID3 wants it, "n/total", dropping the
// total when it is unknown and the frame entirely when the position is: a
// single disc album sends no disc number, so no TPOS is written.
func joinTotal(n, total string) string {
	if n == "" {
		return ""
	}
	if total == "" {
		return n
	}

	return n + "/" + total
}

func addID3Text(tag *id3v2.Tag, name, value string) {
	if value != "" {
		tag.AddTextFrame(name, tag.DefaultEncoding(), value)
	}
}

func addID3Comment(tag *id3v2.Tag, value string) {
	if value != "" {
		tag.AddCommentFrame(id3v2.CommentFrame{
			Encoding: tag.DefaultEncoding(),
			Language: "eng",
			Text:     value,
		})
	}
}

func addID3TXXX(tag *id3v2.Tag, description, value string) {
	if value != "" {
		tag.AddUserDefinedTextFrame(id3v2.UserDefinedTextFrame{
			Encoding:    tag.DefaultEncoding(),
			Description: description,
			Value:       value,
		})
	}
}
