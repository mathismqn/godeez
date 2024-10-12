package tags

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/tags/flac"
	"github.com/mathismqn/godeez/internal/tags/id3v2"
)

func Add(album *deezer.Album, song *deezer.Song, pathFile string) error {
	cover, err := album.GetCoverImage()
	if err != nil {
		return err
	}

	duration, _ := strconv.Atoi(song.Duration)
	song.Duration = fmt.Sprintf("%d", duration*1000)
	dateParts := strings.Split(album.Data.PhysicalReleaseDate, "-")
	if len(dateParts) == 3 {
		album.Data.PhysicalReleaseDate = dateParts[0]
	}

	ext := path.Ext(pathFile)
	if ext == ".mp3" {
		return id3v2.AddTags(album, song, cover, pathFile)
	}

	return flac.AddTags(album, song, cover, pathFile)
}
