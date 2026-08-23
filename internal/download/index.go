package download

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
)

// audioExts is the set of extensions a download can have, derived from the
// same table trackFilename picks from so that a new output format cannot be
// added without the index learning of it. Nothing else under the output
// directory is worth indexing: cover art, file manager clutter and the ledger
// itself are all skipped.
var audioExts = newExtSet()

type indexedFile struct {
	path string
	size int64
}

// fileIndex locates a recorded download that has been moved or renamed, by
// content hash, without reading the whole library.
//
// The walk only stats: a file is opened and hashed once its length matches
// the one being looked for. Size is a coarser fingerprint of the same bytes
// the hash covers, so a mismatch can never hide a file the hash would have
// matched — the hash still decides, the size only spares the reads that were
// always going to fail. Hashes are memoised, so looking up many records in
// one run costs a single pass over the library rather than one pass each.
type fileIndex struct {
	bySize map[int64][]string
	hashes map[string]string
	byHash map[string]indexedFile

	// degraded records that the library was not read in full, which is the
	// difference between "this content is not here" and "this content might
	// be somewhere I could not look".
	degraded bool
}

// newFileIndex walks root and records the length of every audio file.
//
// Unreadable entries and directories are skipped rather than failing the
// walk, since a permission error somewhere in a music library should not
// break the skip check. Only cancellation aborts it. Dotted names are skipped
// whole: they are the ledger, leftover .part files and file manager clutter,
// never something the user moved a download to.
func newFileIndex(ctx context.Context, root string) (*fileIndex, error) {
	index := &fileIndex{
		bySize: make(map[int64][]string),
		hashes: make(map[string]string),
		byHash: make(map[string]indexedFile),
	}

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			index.degraded = true
			return nil
		}

		if path != root && strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if entry.IsDir() || !audioExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			index.degraded = true
			return nil
		}

		index.bySize[info.Size()] = append(index.bySize[info.Size()], path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return index, nil
}

// find returns the path of the file whose content matches hash, along with
// its size so the caller can record it.
//
// A known size narrows the search to the handful of files that could possibly
// match. Zero means the record predates the field, so every file is a
// candidate — the original behaviour, which is why an unmigrated ledger still
// works.
func (f *fileIndex) find(hash string, size int64) (string, int64, bool) {
	if hash == "" {
		return "", 0, false
	}

	// Anything already hashed answers without touching the buckets again.
	if file, ok := f.byHash[hash]; ok {
		return file.path, file.size, true
	}

	if size > 0 {
		if path, ok := f.findIn(f.bySize[size], hash); ok {
			return path, size, true
		}
		return "", 0, false
	}

	for candidateSize, paths := range f.bySize {
		if path, ok := f.findIn(paths, hash); ok {
			return path, candidateSize, true
		}
	}

	return "", 0, false
}

func (f *fileIndex) findIn(paths []string, hash string) (string, bool) {
	for _, path := range paths {
		if f.hashOf(path) == hash {
			return path, true
		}
	}

	return "", false
}

// hashOf hashes path once and remembers the answer in both directions. A file
// that cannot be read is never retried and marks the index degraded: its
// content is unknown rather than absent, so a miss is no longer proof.
func (f *fileIndex) hashOf(path string) string {
	if hash, ok := f.hashes[path]; ok {
		return hash
	}

	hash, size, err := hashFile(path)
	if err != nil {
		f.hashes[path] = ""
		f.degraded = true
		return ""
	}

	f.hashes[path] = hash
	f.byHash[hash] = indexedFile{path: path, size: size}

	return hash
}

// initFileIndex builds the index on first use and reuses it afterwards.
//
// Building it means stat-ing an entire music library, so it is deferred until
// something actually needs it: a run where every recorded path is still valid
// never pays that cost. The error is cached alongside the index so a failed
// build is not retried once per track.
func (d *Downloader) initFileIndex(ctx context.Context) error {
	d.fileIndexOnce.Do(func() {
		d.fileIndex, d.fileIndexErr = newFileIndex(ctx, d.appConfig.OutputDir)
	})

	return d.fileIndexErr
}

func newExtSet() map[string]bool {
	set := map[string]bool{"." + formatExt(""): true}
	for _, ext := range formatExts {
		set["."+ext] = true
	}

	return set
}
