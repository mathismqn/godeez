package update

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/mathismqn/godeez/internal/buildinfo"
	"github.com/mathismqn/godeez/internal/fsutil"
)

const noCheckEnv = "GODEEZ_NO_UPDATE_CHECK"

const (
	// cacheTTL keeps the check to roughly once a day, which is often enough
	// to notice a release without hitting the GitHub API on every command.
	cacheTTL = 24 * time.Hour

	// checkTimeout is deliberately short. The check is a courtesy running
	// alongside a download, so it gives up quickly rather than delaying
	// anything the user actually asked for.
	checkTimeout = 3 * time.Second
)

type cacheEntry struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "godeez", "update.json"), nil
}

// readCache returns the cached result, or false if there is nothing usable.
// Every failure, including a corrupt or unreadable file, is reported the same
// way: the caller simply checks again, so there is nothing to distinguish.
func readCache() (cacheEntry, bool) {
	path, err := cachePath()
	if err != nil {
		return cacheEntry{}, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cacheEntry{}, false
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return cacheEntry{}, false
	}
	if entry.LatestVersion == "" || time.Since(entry.CheckedAt) > cacheTTL {
		return cacheEntry{}, false
	}

	return entry, true
}

func writeCache(version string) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	if err := fsutil.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}

	data, err := json.Marshal(cacheEntry{CheckedAt: time.Now(), LatestVersion: version})
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// check returns the latest version if it is newer than the running one, or
// "" if it is not. The cache is written even when the release turns out not
// to be newer, since the point is to record that GitHub was asked recently,
// and a failure to write it is ignored: an uncacheable check still works, it
// just repeats.
func check(ctx context.Context) (string, error) {
	if entry, ok := readCache(); ok {
		return latestIfNewer(entry.LatestVersion), nil
	}

	release, err := New().Latest(ctx)
	if err != nil {
		return "", err
	}

	latest := release.Version()
	_ = writeCache(latest)

	return latestIfNewer(latest), nil
}

func latestIfNewer(latest string) string {
	if IsNewer(buildinfo.Version(), latest) {
		return latest
	}

	return ""
}

// StartCheck begins a background update check and returns a channel that
// yields the newer version, if there is one, and is closed either way.
//
// It runs concurrently so the check never delays the command the user ran,
// and the channel is buffered so the goroutine exits even if nobody reads the
// result. Errors are swallowed: a failed check is not something to report.
//
// The check is skipped entirely for development builds, which have no version
// to compare, and whenever GODEEZ_NO_UPDATE_CHECK is set, which is the escape
// hatch for packagers and offline use. Both cases close the channel
// immediately so callers need no special handling.
func StartCheck(ctx context.Context) <-chan string {
	ch := make(chan string, 1)

	if os.Getenv(noCheckEnv) != "" || buildinfo.IsDev() {
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)

		ctx, cancel := context.WithTimeout(ctx, checkTimeout)
		defer cancel()

		if latest, err := check(ctx); err == nil && latest != "" {
			ch <- latest
		}
	}()

	return ch
}
