package tags

import (
	"fmt"
	"path"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacvorbis/v2"
	"github.com/go-flac/go-flac/v2"
)

// TrackMetadata holds the metadata extracted from audio files
type TrackMetadata struct {
	Title       string
	Artist      string
	AlbumArtist string
	Album       string
	TrackNumber string
}

// ReadMetadata extracts metadata from an audio file
func ReadMetadata(filePath string) (*TrackMetadata, error) {
	ext := path.Ext(filePath)

	if ext == ".mp3" {
		return readMP3Metadata(filePath)
	} else if ext == ".flac" {
		return readFLACMetadata(filePath)
	}

	return nil, fmt.Errorf("unsupported file format: %s", ext)
}

// readMP3Metadata reads metadata from MP3 files
func readMP3Metadata(filePath string) (*TrackMetadata, error) {
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return nil, err
	}
	defer tag.Close()

	metadata := &TrackMetadata{
		Title:       tag.Title(),
		Artist:      tag.Artist(),
		AlbumArtist: getTextFrame(tag, "TPE2"),
		Album:       tag.Album(),
		TrackNumber: getTextFrame(tag, "TRCK"),
	}

	// Use artist as album artist if album artist is not set
	if metadata.AlbumArtist == "" {
		metadata.AlbumArtist = metadata.Artist
	}

	return metadata, nil
}

// readFLACMetadata reads metadata from FLAC files
func readFLACMetadata(filePath string) (*TrackMetadata, error) {
	file, err := flac.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	cmts, _, err := extractFLACComment(file)
	if err != nil {
		return nil, err
	}

	if cmts == nil {
		return &TrackMetadata{}, nil
	}

	metadata := &TrackMetadata{
		Title:       getVorbisComment(cmts, "TITLE"),
		Artist:      getVorbisComment(cmts, "ARTIST"),
		AlbumArtist: getVorbisComment(cmts, "ALBUMARTIST"),
		Album:       getVorbisComment(cmts, "ALBUM"),
		TrackNumber: getVorbisComment(cmts, "TRACKNUMBER"),
	}

	// Use artist as album artist if album artist is not set
	if metadata.AlbumArtist == "" {
		metadata.AlbumArtist = metadata.Artist
	}

	return metadata, nil
}

// getTextFrame gets a text frame value from ID3v2 tag
func getTextFrame(tag *id3v2.Tag, frameName string) string {
	frames := tag.GetFrames(frameName)
	if len(frames) == 0 {
		return ""
	}

	if textFrame, ok := frames[0].(id3v2.TextFrame); ok {
		return textFrame.Text
	}

	return ""
}

// getVorbisComment gets a comment value from FLAC vorbis comments
func getVorbisComment(cmts *flacvorbis.MetaDataBlockVorbisComment, key string) string {
	comments, _ := cmts.Get(key)
	if len(comments) == 0 {
		return ""
	}
	return comments[0]
}
