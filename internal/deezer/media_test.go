package deezer

import (
	"bytes"
	"encoding/json"
	"testing"
)

const getURLResponse = `{"data":[{"media":[{"format":"FLAC","sources":[{"url":"https://example.invalid/audio"}]}]}]}`

func decodeGetURLResponse(t *testing.T) *mediaResponse {
	t.Helper()

	var res mediaResponse
	if err := json.Unmarshal([]byte(getURLResponse), &res); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	return &res
}

func TestMediaKeyUsesSourceID(t *testing.T) {
	const (
		requested = "2358247065"
		source    = "2134121047"
	)

	got := newMedia(source, decodeGetURLResponse(t)).Key()

	if !bytes.Equal(got, blowfishKey(source)) {
		t.Errorf("Key() is not the key of the track the audio comes from, %s", source)
	}
	if bytes.Equal(got, blowfishKey(requested)) {
		t.Errorf("Key() is the key of the requested track %s, not of the source", requested)
	}
}

func TestNewMedia(t *testing.T) {
	media := newMedia("2134121047", decodeGetURLResponse(t))

	if media.sourceID != "2134121047" {
		t.Errorf("sourceID = %q, want 2134121047", media.sourceID)
	}
	if media.Format() != "FLAC" {
		t.Errorf("Format() = %q, want FLAC", media.Format())
	}
	if media.url != "https://example.invalid/audio" {
		t.Errorf("url = %q, want https://example.invalid/audio", media.url)
	}
}
