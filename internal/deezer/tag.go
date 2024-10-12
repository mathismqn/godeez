package deezer

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/bogem/id3v2/v2"
)

func AddTags(album *Album, song *Song, pathFile string) error {
	cover, err := getCoverImage(album.Data.CoverID)
	if err != nil {
		return err
	}

	ext := path.Ext(pathFile)
	if ext == ".mp3" {
		tag, err := id3v2.Open(pathFile, id3v2.Options{Parse: true})
		if err != nil {
			return err
		}
		defer tag.Close()

		var year string
		duration, _ := strconv.Atoi(song.Duration)
		dateParts := strings.Split(album.Data.PhysicalReleaseDate, "-")
		if len(dateParts) == 3 {
			year = dateParts[0]
		}

		addTag(tag, "TALB", album.Data.Name)
		addTag(tag, "TPE2", album.Data.Artist)
		addTag(tag, "TPUB", album.Data.Label)
		addTag(tag, "TDOR", album.Data.OriginalReleaseDate)
		addTag(tag, "COMM", album.Data.ProducerLine)
		if year != "" {
			addTag(tag, "TYER", year)
		}

		addTag(tag, "TIT2", song.Title)
		addTag(tag, "TPE1", strings.Join(song.Contributors.MainArtists, " / "))
		addTag(tag, "TCOM", strings.Join(song.Contributors.Composers, " / "))
		addTag(tag, "TEXT", strings.Join(song.Contributors.Authors, " / "))
		addTag(tag, "TRCK", song.TrackNumber)
		if duration > 0 {
			addTag(tag, "TLEN", fmt.Sprintf("%d", duration*1000))
		}
		addTXXXTag(tag, "GAIN", song.Gain)
		addTXXXTag(tag, "ISRC", song.ISRC)
		addTXXXTag(tag, "Explicit lyrics", song.ExplicitLyrics)

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

	return nil
}

func getCoverImage(id string) ([]byte, error) {
	url := fmt.Sprintf("https://e-cdn-images.dzcdn.net/images/cover/%s/500x500-000000-80-0-0.jpg", id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
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
