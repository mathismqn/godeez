package tag

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-flac/flacvorbis/v2"
	goflac "github.com/go-flac/go-flac/v2"
	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
	"github.com/mewkiz/flac/meta"
)

// writeTestFLAC creates a minimal, valid flac file, mirroring what an encoder
// produces: a STREAMINFO block followed by the audio frames.
func writeTestFLAC(t *testing.T, path string) {
	t.Helper()

	info := &meta.StreamInfo{
		BlockSizeMin:  16,
		BlockSizeMax:  16,
		SampleRate:    44100,
		NChannels:     1,
		BitsPerSample: 16,
		NSamples:      16,
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer f.Close()

	enc, err := flac.NewEncoder(f, info)
	if err != nil {
		t.Fatalf("encoder: %v", err)
	}

	subframe := &frame.Subframe{
		SubHeader: frame.SubHeader{Pred: frame.PredVerbatim},
		Samples:   make([]int32, 16),
		NSamples:  16,
	}
	fl := &frame.Frame{
		Header: frame.Header{
			HasFixedBlockSize: true,
			BlockSize:         16,
			SampleRate:        44100,
			Channels:          frame.ChannelsMono,
			BitsPerSample:     16,
		},
		Subframes: []*frame.Subframe{subframe},
	}
	if err := enc.WriteFrame(fl); err != nil {
		t.Fatalf("write frame: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("close encoder: %v", err)
	}
}

// addVorbisComments gives a test file the comment block Deezer ships some
// downloads with.
func addVorbisComments(t *testing.T, path string, comments [][2]string) {
	t.Helper()

	f, err := goflac.ParseFile(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cmts := flacvorbis.New()
	for _, c := range comments {
		if err := cmts.Add(c[0], c[1]); err != nil {
			t.Fatal(err)
		}
	}
	commentMeta := cmts.Marshal()
	f.Meta = append(f.Meta, &commentMeta)
	if err := f.Save(path + ".new"); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := os.Rename(path+".new", path); err != nil {
		t.Fatal(err)
	}
}

// readVorbis decodes the Vorbis comments of path with an independent reader,
// so it reports what a third party tool would see rather than what the writer
// believes it wrote. Repeated names keep every value.
func readVorbis(t *testing.T, path string) map[string][]string {
	t.Helper()

	stream, err := flac.ParseFile(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	defer stream.Close()

	out := map[string][]string{}
	for _, block := range stream.Blocks {
		if block.Header.Type != meta.TypeVorbisComment {
			continue
		}
		cmts, ok := block.Body.(*meta.VorbisComment)
		if !ok {
			t.Fatalf("vorbis comment body is %T", block.Body)
		}
		for _, kv := range cmts.Tags {
			out[kv[0]] = append(out[kv[0]], kv[1])
		}
	}

	return out
}

// TestWriteFLACTags checks that the tags the writer is asked for reach the
// Vorbis comments, read back with a parser from a different library than the
// one that wrote them.
func TestWriteFLACTags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "track.flac")
	writeTestFLAC(t, path)

	m := Metadata{
		Title:       "Song",
		Artists:     "The Artist",
		TrackNumber: "1",
		Key:         "Am",
		Album: &AlbumMetadata{
			Artist:      "The Artist",
			Title:       "The Album",
			ReleaseDate: "2020-05-01",
		},
	}
	if err := Write(path, m); err != nil {
		t.Fatalf("write: %v", err)
	}

	got := readVorbis(t, path)
	for key, want := range map[string]string{
		"ARTIST":      "The Artist",
		"ALBUM":       "The Album",
		"ALBUMARTIST": "The Artist",
		"TITLE":       "Song",
		"TRACKNUMBER": "1",
		"DATE":        "2020",
		"KEY":         "Am",
		"INITIALKEY":  "Am",
	} {
		if values := got[key]; len(values) != 1 || values[0] != want {
			t.Errorf("%s = %v, want a single %q", key, values, want)
		}
	}
}

// TestWriteFLACOverDeezerTags tags a file shaped like the ones Deezer serves
// for parts of its older catalogue: already tagged, with a title that lacks
// the version, an empty artist and an empty copyright. The values godeez
// writes have to take the place of those rather than pile up beside them.
func TestWriteFLACOverDeezerTags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "track.flac")
	writeTestFLAC(t, path)
	addVorbisComments(t, path, [][2]string{
		{"TITLE", "Save the World"},
		{"artist", ""},
		{"COPYRIGHT", ""},
		{"ALBUMTITLE", "Save The World"},
		{"COMMENT", "5099908299352"},
	})

	m := Metadata{
		Title:   "Save The World (Extended Mix)",
		Artists: "Swedish House Mafia",
		Album:   &AlbumMetadata{Title: "Save The World"},
	}
	if err := Write(path, m); err != nil {
		t.Fatalf("write: %v", err)
	}

	got := readVorbis(t, path)
	for key, want := range map[string]string{
		// Written by godeez: one current value, whatever the file had.
		"TITLE":  "Save The World (Extended Mix)",
		"ARTIST": "Swedish House Mafia",
		"ALBUM":  "Save The World",
		// Only Deezer writes these: kept as they were.
		"ALBUMTITLE": "Save The World",
		"COMMENT":    "5099908299352",
	} {
		if values := got[key]; len(values) != 1 || values[0] != want {
			t.Errorf("%s = %v, want a single %q", key, values, want)
		}
	}
	if values := got["artist"]; len(values) != 0 {
		t.Errorf("artist = %v, want the empty lowercase copy dropped", values)
	}
	if values := got["COPYRIGHT"]; len(values) != 0 {
		t.Errorf("COPYRIGHT = %v, want the empty comment dropped", values)
	}
}
