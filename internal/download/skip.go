package download

import (
	"context"

	"github.com/mathismqn/godeez/internal/fsutil"
)

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
