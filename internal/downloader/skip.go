package downloader

import (
	"context"

	"github.com/mathismqn/godeez/internal/fileutil"
	"github.com/mathismqn/godeez/internal/store"
)

func (c *Client) shouldSkipDownload(ctx context.Context, songID, mediaFormat string) (string, bool) {
	existing, err := store.GetDownloadInfo(songID)
	if err != nil || existing.Quality != mediaFormat {
		return "", false
	}

	if fileutil.FileExists(existing.Path) {
		return existing.Path, true
	}

	if existing.Hash == "" {
		return "", false
	}

	if err := c.initHashIndex(ctx); err != nil {
		return "", false
	}

	foundPath, ok := c.hashIndex.Find(existing.Hash)
	if !ok {
		return "", false
	}

	existing.Path = foundPath
	_ = existing.Save()

	return foundPath, true
}
