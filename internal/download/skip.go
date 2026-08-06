package download

import (
	"context"

	"github.com/mathismqn/godeez/internal/fsutil"
)

// shouldSkipDownload reports whether trackID has already been downloaded at
// mediaFormat, returning the path of the existing file.
//
// A recorded download at a different quality is not a skip: asking for flac
// after previously fetching mp3_128 should download again.
//
// When the recorded path is gone the file may simply have been moved or
// renamed by the user, so the content hash is used to look for it elsewhere
// under the output directory before giving up. A match repairs the stored
// path, which keeps the ledger useful across library reorganisations. That
// lookup is best effort throughout: every failure falls through to
// downloading again, which is always safe.
func (d *Downloader) shouldSkipDownload(ctx context.Context, trackID, mediaFormat string) (string, bool) {
	existing, err := d.store.DownloadInfo(trackID)
	if err != nil || existing.Quality != mediaFormat {
		return "", false
	}

	if fsutil.Exists(existing.Path) {
		return existing.Path, true
	}

	if existing.Hash == "" {
		return "", false
	}

	if err := d.initHashIndex(ctx); err != nil {
		return "", false
	}

	foundPath, ok := d.hashIndex.find(existing.Hash)
	if !ok {
		return "", false
	}

	existing.Path = foundPath
	_ = d.store.PutDownloadInfo(existing)

	return foundPath, true
}
