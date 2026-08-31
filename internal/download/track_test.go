package download

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mathismqn/godeez/internal/deezer"
)

// TestDownloadTrackSkipsPersonalUpload builds the Downloader without a Deezer
// client on purpose: reaching FetchMedia would panic, so the check is pinned
// to the top of downloadTrack rather than merely somewhere inside it.
func TestDownloadTrackSkipsPersonalUpload(t *testing.T) {
	dir := t.TempDir()
	d := newTestDownloader(t, dir)

	track := &deezer.Track{ID: "-3002903542", Type: "1", Title: "01 Omotesando", Artist: "Népal"}
	result := d.downloadTrack(context.Background(), nil, track, Options{}, dir, nil)

	if result.skip != skipPersonalUpload {
		t.Fatalf("skip = %v, want skipPersonalUpload (err = %v)", result.skip, result.err)
	}
	if result.err != nil {
		t.Errorf("err = %v, want nil", result.err)
	}
}

// roundTripFunc serves a canned response so a test can drive the real
// FetchMedia rather than a stand-in for it.
type roundTripFunc func(*http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req), nil }

// TestDownloadTrackSkipsUnavailable is the outcome that has to survive the
// whole resolver: a track Deezer refuses, and publishes no stand-in for, is a
// skip rather than a failure. The media server answers with the 2002 it sends
// for an entry carrying no streaming rights.
func TestDownloadTrackSkipsUnavailable(t *testing.T) {
	dir := t.TempDir()
	d := newTestDownloader(t, dir)
	d.deezerClient = &deezer.Client{
		Session: &deezer.Session{
			HTTPClient: &http.Client{
				Transport: roundTripFunc(func(*http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"data":[{"media":[],"errors":[{"code":2002,"message":"no rights"}]}]}`)),
						Header:     make(http.Header),
					}
				}),
			},
		},
	}

	track := &deezer.Track{ID: "3059676021", Title: "Remember Me", Artist: "d4vd", TrackToken: "token"}
	result := d.downloadTrack(context.Background(), nil, track, Options{Quality: "mp3_320"}, dir, nil)

	if result.skip != skipUnavailable {
		t.Fatalf("skip = %v, want skipUnavailable (err = %v)", result.skip, result.err)
	}
	if result.err != nil {
		t.Errorf("err = %v, want nil", result.err)
	}
}
