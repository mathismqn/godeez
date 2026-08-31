// Package tag writes track metadata into finished audio files.
//
// Every container stores metadata differently: mp3 uses ID3v2 frames, flac
// uses Vorbis comments, and wav carries an ID3 chunk plus a RIFF LIST/INFO
// chunk for players that read only one of the two. Write hides that behind a
// single Metadata struct and dispatches on the file extension.
//
// The taggers are written for freshly downloaded files. Empty fields are
// skipped rather than written as blanks, and each tagger writes through a
// temporary file so a failure part way cannot corrupt the audio.
package tag

import (
	"path/filepath"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
)

// AlbumMetadata is the subset of tags that only make sense for a track that
// belongs to an album. It is nil on a standalone single.
type AlbumMetadata struct {
	Artist              string
	Title               string
	Label               string
	OriginalReleaseDate string
	ReleaseDate         string
	ProducerLine        string
	Copyright           string
}

// Metadata is the container-independent tag set. Every field is a string
// because the underlying formats store them as text; conversions such as
// Duration to milliseconds happen inside the individual taggers.
type Metadata struct {
	Title       string
	Artists     string
	Composers   string
	Lyricists   string
	Genre       string
	BPM         string
	Key         string
	TrackNumber string
	TrackTotal  string
	DiscNumber  string
	DiscTotal   string
	Duration    string
	Gain        string
	ISRC        string
	Cover       []byte
	Album       *AlbumMetadata
}

type tagger interface {
	write(m Metadata) error
}

// newTagger picks an implementation from the file extension. Anything that is
// not mp3 or wav is attempted as flac rather than rejected, so an unexpected
// extension fails with a parse error from the flac library instead of a
// generic unsupported-format message.
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
