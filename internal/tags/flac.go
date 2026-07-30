package tags

import (
	"os"
	"strings"

	"github.com/go-flac/flacpicture/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
	"github.com/mathismqn/godeez/internal/deezer"
)

type flacTagger struct {
	file  *flac.File
	cmts  *flacvorbis.MetaDataBlockVorbisComment
	index int
}

func (t *flacTagger) addTags(resource deezer.Resource, track *deezer.Track, cover []byte, path, tempo, key, genre string) error {
	if album, ok := resource.(*deezer.Album); ok {
		if parts := strings.Split(album.Results.Data.PhysicalReleaseDate, "-"); len(parts) == 3 {
			album.Results.Data.PhysicalReleaseDate = parts[0]
		}

		t.addTag("TRACKNUMBER", track.TrackNumber)
		t.addTag("ALBUMARTIST", album.Results.Data.Artist)
		t.addTag("ALBUM", album.Results.Data.Title)
		t.addTag("PUBLISHER", album.Results.Data.Label)
		t.addTag("ORIGINALDATE", album.Results.Data.OriginalReleaseDate)
		t.addTag("DATE", album.Results.Data.PhysicalReleaseDate)
		t.addTag("COMMENT", album.Results.Data.ProducerLine)
		t.addTag("COPYRIGHT", album.Results.Data.Copyright)
	}

	t.addTag("ARTIST", strings.Join(track.Contributors.MainArtists, ", "))
	t.addTag("TITLE", track.GetTitle())
	t.addTag("COMPOSER", strings.Join(track.Contributors.Composers, ", "))
	t.addTag("LYRICIST", strings.Join(track.Contributors.Authors, ", "))
	t.addTag("GENRE", genre)
	t.addTag("REPLAYGAIN_TRACK_GAIN", track.Gain)
	t.addTag("ISRC", track.ISRC)
	t.addTag("BPM", tempo)
	t.addTag("KEY", key)
	t.addTag("INITIALKEY", key)

	cmtsMeta := t.cmts.Marshal()
	if t.index > 0 {
		t.file.Meta[t.index] = &cmtsMeta
	} else {
		t.file.Meta = append(t.file.Meta, &cmtsMeta)
	}

	picture, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "Front cover", cover, "image/jpeg")
	if err != nil {
		return err
	}
	pictureMeta := picture.Marshal()
	t.file.Meta = append(t.file.Meta, &pictureMeta)

	tmpPath := path + ".tmp"
	if err := t.file.Save(tmpPath); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
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
