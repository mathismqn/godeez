package tags

import (
	"path"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
	"github.com/mathismqn/godeez/internal/deezer"
)

type tagger interface {
	addTags(resource deezer.Resource, song *deezer.Song, cover []byte, path, tempo, key, genre string) error
}

func newTagger(filePath string) (tagger, error) {
	ext := path.Ext(filePath)
	if ext == ".mp3" {
		tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
		if err != nil {
			return nil, err
		}

		return &id3v2Tagger{tag: tag}, nil
	}

	file, err := flac.ParseFile(filePath)
	if err != nil {
		return nil, err
	}
	cmts, idx, err := extractFLACComment(file)
	if err != nil {
		return nil, err
	}
	if cmts == nil && idx > 0 {
		cmts = flacvorbis.New()
	}

	return &flacTagger{file: file, cmts: cmts, index: idx}, nil
}

func AddTags(resource deezer.Resource, song *deezer.Song, cover []byte, filePath, tempo, key, genre string) error {
	tagger, err := newTagger(filePath)
	if err != nil {
		return err
	}

	return tagger.addTags(resource, song, cover, filePath, tempo, key, genre)
}
