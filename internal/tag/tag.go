package tag

import (
	"path/filepath"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
)

type AlbumMetadata struct {
	Artist              string
	Title               string
	Label               string
	OriginalReleaseDate string
	ReleaseDate         string
	ProducerLine        string
	Copyright           string
}

type Metadata struct {
	Title       string
	Artists     string
	Composers   string
	Lyricists   string
	Genre       string
	BPM         string
	Key         string
	TrackNumber string
	Duration    string
	Gain        string
	ISRC        string
	Cover       []byte
	Album       *AlbumMetadata
}

type tagger interface {
	write(m Metadata) error
}

func newTagger(filePath string) (tagger, error) {
	switch filepath.Ext(filePath) {
	case ".mp3":
		tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
		if err != nil {
			return nil, err
		}
		return &id3v2Tagger{tag: tag}, nil
	case ".wav":
		return &wavTagger{path: filePath}, nil
	}

	file, err := flac.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	cmts, idx := extractFLACComment(file)
	if cmts == nil {
		cmts = flacvorbis.New()
	}

	return &flacTagger{file: file, cmts: cmts, index: idx, path: filePath}, nil
}

func Write(filePath string, m Metadata) error {
	t, err := newTagger(filePath)
	if err != nil {
		return err
	}
	return t.write(m)
}
