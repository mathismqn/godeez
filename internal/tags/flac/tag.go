package flac

import (
	"os"
	"strings"

	"github.com/go-flac/flacpicture/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
	"github.com/mathismqn/godeez/internal/deezer"
)

func AddTags(album *deezer.Album, song *deezer.Song, cover []byte, path string) error {
	file, err := flac.ParseFile(path)
	if err != nil {
		return err
	}
	cmts, idx, err := extractFLACComment(file)
	if err != nil {
		return err
	}
	if cmts == nil && idx > 0 {
		cmts = flacvorbis.New()
	}

	addTag(cmts, "ALBUM", album.Data.Name)
	addTag(cmts, "ALBUMARTIST", album.Data.Artist)
	addTag(cmts, "PUBLISHER", album.Data.Label)
	addTag(cmts, "ORIGINALDATE", album.Data.OriginalReleaseDate)
	addTag(cmts, "DATE", album.Data.PhysicalReleaseDate)
	addTag(cmts, "COMMENT", album.Data.ProducerLine)

	addTag(cmts, "TITLE", song.Title)
	addTag(cmts, "ARTIST", strings.Join(song.Contributors.MainArtists, " / "))
	addTag(cmts, "COMPOSER", strings.Join(song.Contributors.Composers, " / "))
	addTag(cmts, "LYRICIST", strings.Join(song.Contributors.Authors, " / "))
	addTag(cmts, "TRACKNUMBER", song.TrackNumber)
	addTag(cmts, "REPLAYGAIN_TRACK_GAIN", song.Gain)
	addTag(cmts, "ISRC", song.ISRC)

	cmtsmeta := cmts.Marshal()
	if idx > 0 {
		file.Meta[idx] = &cmtsmeta
	} else {
		file.Meta = append(file.Meta, &cmtsmeta)
	}

	picture, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "Front cover", cover, "image/jpeg")
	if err != nil {
		return err
	}
	picturemeta := picture.Marshal()
	file.Meta = append(file.Meta, &picturemeta)

	return saveTags(file, path)
}

func addTag(cmts *flacvorbis.MetaDataBlockVorbisComment, name, value string) {
	if value != "" {
		cmts.Add(name, value)
	}
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

func saveTags(file *flac.File, path string) error {
	tempPath := path + ".tmp"
	file.Save(tempPath)

	return os.Rename(tempPath, path)
}
