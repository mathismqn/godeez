package download

import (
	"context"
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
