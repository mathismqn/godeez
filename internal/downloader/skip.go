package downloader

import (
	"context"

	"github.com/mathismqn/godeez/internal/fileutil"
	"github.com/mathismqn/godeez/internal/store"
)

type SkipError struct {
	Path string
}

func (e SkipError) Error() string {
	return e.Path
}

func IsSkipError(err error) (string, bool) {
	if skipErr, ok := err.(SkipError); ok {
		return skipErr.Path, true
	}
	return "", false
}

func (c *Client) shouldSkipDownload(ctx context.Context, songID, mediaFormat string) (string, bool) {
	if existing, err := store.GetDownloadInfo(songID); err == nil && existing.Quality == mediaFormat {
		if fileutil.FileExists(existing.Path) {
			return existing.Path, true
		}
		if existing.Hash != "" {
			if err := c.initHashIndex(ctx); err == nil {
				if foundPath, ok := c.hashIndex.Find(existing.Hash); ok {
					existing.Path = foundPath
					_ = existing.Save()

					return foundPath, true
				}
			}
		}
	}

	return "", false
}
