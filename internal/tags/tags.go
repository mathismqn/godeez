package tags

import (
	"path"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
	"github.com/mathismqn/godeez/internal/deezer"
)

type tagger interface {
	addTags(resource deezer.Resource, track *deezer.Track, cover []byte, filePath, tempo, key, genre string) error
}

func newTagger(filePath string) (tagger, error) {
	if path.Ext(filePath) == ".mp3" {
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

	cmts, idx := extractFLACComment(file)
	if cmts == nil {
		cmts = flacvorbis.New()
	}

	return &flacTagger{file: file, cmts: cmts, index: idx}, nil
}

func AddTags(resource deezer.Resource, track *deezer.Track, cover []byte, filePath, tempo, key, genre string) error {
	t, err := newTagger(filePath)
	if err != nil {
		return err
	}
	return t.addTags(resource, track, cover, filePath, tempo, key, genre)
}
