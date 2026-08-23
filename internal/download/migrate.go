package download

import (
	"context"
	"os"

	"github.com/mathismqn/godeez/internal/store"
)

// sizeSchema is the ledger version this migration produces. A later migration
// declares the next number and guards on its own, so that adding one cannot
// make this one claim to have done its work.
const sizeSchema = 1

// migrateSizes fills in the size of every record written before the ledger
// tracked one, so that later runs can find a moved file by stat-ing
// candidates instead of reading the whole library.
//
// It runs at most once. The version marker is what guarantees that: a record
// whose file was deleted can never be given a size, so without the marker
// every future run would rediscover it and pay for the escalation again.
//
// The work is best effort, in the same spirit as MigrateLegacy: a ledger that
// keeps its old records is merely slow, never wrong.
func (d *Downloader) migrateSizes(ctx context.Context) {
	if d.store.Version() >= sizeSchema {
		return
	}

	updated, dead, complete := d.resolveSizes(ctx)

	// Whatever was resolved is worth keeping even on a partial run: those
	// records now have a size and are skipped by the next attempt.
	if err := d.store.UpdateDownloadInfos(updated, dead); err != nil {
		return
	}

	if complete {
		_ = d.store.SetVersion(sizeSchema)
	}
}

// resolveSizes returns the records it could measure, the ids of the records
// it should drop, and whether it got through all of them.
//
// Most records are answered by a stat of the path they already point at. Only
// records whose file has since moved need the index, and building it reads
// every file in the library, so a ledger where nothing has moved never
// touches it.
func (d *Downloader) resolveSizes(ctx context.Context) (updated []*store.DownloadInfo, dead []string, complete bool) {
	// Collected first, stat-ed after: the callback runs inside a read
	// transaction that a large ledger would otherwise hold open across every
	// one of those filesystem calls.
	var unsized []*store.DownloadInfo
	if err := d.store.EachDownloadInfo(func(info *store.DownloadInfo) error {
		if info.Size == 0 {
			unsized = append(unsized, info)
		}
		return nil
	}); err != nil {
		return nil, nil, false
	}

	var moved []*store.DownloadInfo
	for _, info := range unsized {
		if ctx.Err() != nil {
			return updated, nil, false
		}

		if stat, err := os.Stat(info.Path); err == nil && !stat.IsDir() {
			info.Size = stat.Size()
			updated = append(updated, info)
			continue
		}

		moved = append(moved, info)
	}

	if len(moved) == 0 {
		return updated, nil, true
	}

	// Everything from here reads the library rather than stat-ing it, so the
	// user is told what the wait is for.
	sp := startLedgerScan()
	defer finishLedgerScan(sp)

	if err := d.initFileIndex(ctx); err != nil {
		return updated, nil, false
	}

	// No audio under the output directory at all means the library was not
	// where it was expected rather than emptied, and nothing here can match.
	if len(d.fileIndex.bySize) == 0 {
		return updated, nil, true
	}

	for _, info := range moved {
		if ctx.Err() != nil {
			return updated, d.prunable(dead), false
		}

		path, size, ok := d.fileIndex.find(info.Hash, 0)
		if !ok {
			// Neither its path nor its content is anywhere in the library, so
			// the record describes nothing and no reader can do more with it
			// than pay to rediscover that.
			dead = append(dead, info.TrackID)
			continue
		}

		info.Path = path
		info.Size = size
		updated = append(updated, info)
	}

	return updated, d.prunable(dead), true
}

// prunable filters the dead list down to what the index is entitled to
// condemn.
//
// A miss only proves a file is gone if the library was read in full. Where
// part of it could not be listed or opened, a file that is merely out of
// sight looks exactly like one that was deleted, and its record is the only
// thing that can recover it once the problem is fixed.
func (d *Downloader) prunable(dead []string) []string {
	if d.fileIndex.degraded {
		return nil
	}

	return dead
}
