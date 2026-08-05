package downloader

import (
	"context"

	"github.com/mathismqn/godeez/internal/fsutil"
)

func (c *Client) shouldSkipDownload(ctx context.Context, trackID, mediaFormat string) (string, bool) {
	existing, err := c.store.DownloadInfo(trackID)
	if err != nil || existing.Quality != mediaFormat {
		return "", false
	}

	if fsutil.Exists(existing.Path) {
		return existing.Path, true
	}

	if existing.Hash == "" {
		return "", false
	}

	if err := c.initHashIndex(ctx); err != nil {
		return "", false
	}

	foundPath, ok := c.hashIndex.find(existing.Hash)
	if !ok {
		return "", false
	}

	existing.Path = foundPath
	_ = c.store.PutDownloadInfo(existing)

	return foundPath, true
}
