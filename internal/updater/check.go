package updater

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/mathismqn/godeez/internal/buildinfo"
	"github.com/mathismqn/godeez/internal/fileutil"
)

const noCheckEnv = "GODEEZ_NO_UPDATE_CHECK"

const (
	cacheTTL     = 24 * time.Hour
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
	if err := fileutil.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}

	data, err := json.Marshal(cacheEntry{CheckedAt: time.Now(), LatestVersion: version})
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func check(ctx context.Context) (string, error) {
	if entry, ok := readCache(); ok {
		return newerThanCurrent(entry.LatestVersion), nil
	}

	release, err := New().Latest(ctx)
	if err != nil {
		return "", err
	}

	latest := release.Version()
	_ = writeCache(latest)

	return newerThanCurrent(latest), nil
}

func newerThanCurrent(latest string) string {
	if IsNewer(buildinfo.Version(), latest) {
		return latest
	}

	return ""
}

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
