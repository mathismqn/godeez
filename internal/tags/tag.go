package tags

import (
	"path"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
	"github.com/mathismqn/godeez/internal/deezer"
)

type Tagger interface {
	AddTags(resource deezer.Resource, song *deezer.Song, cover []byte, path, tempo, key string) error
}

func NewTagger(filePath string) (Tagger, error) {
	ext := path.Ext(filePath)
	if ext == ".mp3" {
		tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
		if err != nil {
			return nil, err
		}
		return &ID3v2Tagger{Tag: tag}, nil
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

	return &FLACTagger{File: file, Cmts: cmts, Index: idx}, nil
}

func AddTags(resource deezer.Resource, song *deezer.Song, filePath, tempo, key string) error {
	tagger, err := NewTagger(filePath)
	if err != nil {
		return err
	}
	cover, err := song.GetCoverImage()
	if err != nil {
		return err
	}

	return tagger.AddTags(resource, song, cover, filePath, tempo, key)
}
