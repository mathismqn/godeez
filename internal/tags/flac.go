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

func (t *flacTagger) addTags(resource deezer.Resource, song *deezer.Song, cover []byte, path, tempo, key string) error {
	if album, ok := resource.(*deezer.Album); ok {
		dateParts := strings.Split(album.Results.Data.PhysicalReleaseDate, "-")
		if len(dateParts) == 3 {
			album.Results.Data.PhysicalReleaseDate = dateParts[0]
		}

		t.addTag("TRACKNUMBER", song.TrackNumber)
		t.addTag("ALBUMARTIST", album.Results.Data.Artist)
		t.addTag("ALBUM", album.Results.Data.Title)
		t.addTag("PUBLISHER", album.Results.Data.Label)
		t.addTag("ORIGINALDATE", album.Results.Data.OriginalReleaseDate)
		t.addTag("DATE", album.Results.Data.PhysicalReleaseDate)
		t.addTag("COMMENT", album.Results.Data.ProducerLine)
		t.addTag("COPYRIGHT", album.Results.Data.Copyright)
	}

	t.addTag("ARTIST", strings.Join(song.Contributors.MainArtists, ", "))
	t.addTag("TITLE", song.Title)
	t.addTag("COMPOSER", strings.Join(song.Contributors.Composers, ", "))
	t.addTag("LYRICIST", strings.Join(song.Contributors.Authors, ", "))
	t.addTag("REPLAYGAIN_TRACK_GAIN", song.Gain)
	t.addTag("ISRC", song.ISRC)

	t.addTag("BPM", tempo)
	t.addTag("KEY", key)
	t.addTag("INITIALKEY", key)

	cmtsmeta := t.cmts.Marshal()
	if t.index > 0 {
		t.file.Meta[t.index] = &cmtsmeta
	} else {
		t.file.Meta = append(t.file.Meta, &cmtsmeta)
	}

	picture, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "Front cover", cover, "image/jpeg")
	if err != nil {
		return err
	}
	picturemeta := picture.Marshal()
	t.file.Meta = append(t.file.Meta, &picturemeta)

	return t.saveTags(path)
}

func (t *flacTagger) addTag(name, value string) {
	if value != "" {
		t.cmts.Add(name, value)
	}
}

func (t *flacTagger) saveTags(path string) error {
	tempPath := path + ".tmp"
	t.file.Save(tempPath)

	return os.Rename(tempPath, path)
}

func extractFLACComment(file *flac.File) (*flacvorbis.MetaDataBlockVorbisComment, int, error) {
	var cmt *flacvorbis.MetaDataBlockVorbisComment
	var cmtIdx int
	var err error
	for idx, meta := range file.Meta {
		if meta.Type == flac.VorbisComment {
			cmt, err = flacvorbis.ParseFromMetaDataBlock(*meta)
			cmtIdx = idx
			if err != nil {
				return nil, 0, err
			}
		}
	}

	return cmt, cmtIdx, nil
}
