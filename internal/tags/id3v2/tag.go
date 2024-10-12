package id3v2

import (
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/mathismqn/godeez/internal/deezer"
)

func AddTags(album *deezer.Album, song *deezer.Song, cover []byte, path string) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()

	addTag(tag, "TALB", album.Data.Name)
	addTag(tag, "TPE2", album.Data.Artist)
	addTag(tag, "TPUB", album.Data.Label)
	addTag(tag, "TDOR", album.Data.OriginalReleaseDate)
	addTag(tag, "TYER", album.Data.PhysicalReleaseDate)
	addTag(tag, "COMM", album.Data.ProducerLine)

	addTag(tag, "TIT2", song.Title)
	addTag(tag, "TPE1", strings.Join(song.Contributors.MainArtists, " / "))
	addTag(tag, "TCOM", strings.Join(song.Contributors.Composers, " / "))
	addTag(tag, "TEXT", strings.Join(song.Contributors.Authors, " / "))
	addTag(tag, "TRCK", song.TrackNumber)
	addTag(tag, "TLEN", song.Duration)
	addTXXXTag(tag, "GAIN", song.Gain)
	addTXXXTag(tag, "ISRC", song.ISRC)

	frame := id3v2.PictureFrame{
		Encoding:    tag.DefaultEncoding(),
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Cover",
		Picture:     cover,
	}
	tag.AddAttachedPicture(frame)

	return tag.Save()
}

func addTag(tag *id3v2.Tag, name, value string) {
	if value != "" {
		tag.AddTextFrame(name, tag.DefaultEncoding(), value)
	}
}

func addTXXXTag(tag *id3v2.Tag, description, value string) {
	if value != "" {
		udf := id3v2.UserDefinedTextFrame{
			Encoding:    tag.DefaultEncoding(),
			Description: description,
			Value:       value,
		}
		tag.AddUserDefinedTextFrame(udf)
	}
}
