package download

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mathismqn/godeez/internal/audio"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/fsutil"
	"github.com/mathismqn/godeez/internal/store"
	"github.com/mathismqn/godeez/internal/tag"
)

// downloadTrack runs the whole pipeline for one track and reports the outcome
// rather than returning an error, so the caller can keep going.
//
// Ordering matters here. The format is resolved before the skip check,
// because whether a track counts as already downloaded depends on the format
// that will actually be written, which is not always the one requested. The
// external metadata lookup is started concurrently and collected late, since
// it hits third party sites and is the slowest part of the pipeline while
// also being the least important. Tagging and the store write happen last, in
// finalizeDownload, once the file is known to be complete.
//
// Only cancellation and a failure to produce the audio itself are fatal.
// Everything else, including a missing cover or a quality downgrade, is
// reported as a warning.
func (d *Downloader) downloadTrack(ctx context.Context, resource deezer.Resource, track *deezer.Track, opts Options, outputDir string) downloadResult {
	media, err := d.deezerClient.FetchMedia(ctx, track, opts.sourceQuality())
	if err != nil {
		return downloadResult{err: fmt.Errorf("failed to fetch media: %w", err)}
	}

	mediaFormat := media.Format()
	outputFormat := mediaFormat
	if opts.convertsToWAV() {
		if mediaFormat != "FLAC" {
			return downloadResult{err: fmt.Errorf("wav requires a flac source, but only '%s' is available", strings.ToLower(mediaFormat))}
		}
		outputFormat = "WAV"
	}

	if opts.Strict && strings.ToLower(outputFormat) != opts.Quality {
		return downloadResult{err: fmt.Errorf("requested quality '%s' not available", opts.Quality)}
	}

	if skipPath, skip := d.shouldSkipDownload(ctx, track.ID, outputFormat); skip {
		return downloadResult{skipped: true, path: skipPath}
	}

	metadataChan := make(chan metadataResult, 1)
	go func() {
		metadataChan <- fetchMetadata(ctx, d.deezerClient.Session.HTTPClient, track, opts)
	}()

	dlCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	stream, err := d.deezerClient.MediaStream(dlCtx, media)
	if err != nil {
		return downloadResult{err: fmt.Errorf("failed to get media stream: %w", err)}
	}

	fileName := trackFilename(track, d.kind, outputFormat)
	outputPath := d.uniqueOutputPath(track.ID, filepath.Join(outputDir, fileName))
	key := media.Key()

	if opts.convertsToWAV() {
		tmpPath, err := d.streamToTempFile(dlCtx, stream, outputDir, key)
		if err != nil {
			return downloadResult{err: fmt.Errorf("failed to stream to file: %w", err)}
		}
		defer fsutil.Remove(tmpPath)

		if err := audio.FLACToWAV(ctx, tmpPath, outputPath); err != nil {
			return downloadResult{err: fmt.Errorf("failed to convert to wav: %w", err)}
		}
	} else if err := d.streamToFile(dlCtx, stream, outputPath, key); err != nil {
		return downloadResult{err: fmt.Errorf("failed to stream to file: %w", err)}
	}

	var warnings []string

	if opts.Quality != strings.ToLower(outputFormat) {
		warnings = append(warnings, fmt.Sprintf("requested quality '%s' not available, using '%s' instead", opts.Quality, strings.ToLower(outputFormat)))
	}

	cover, err := d.deezerClient.FetchCoverImage(ctx, track)
	if err != nil && !errors.Is(err, context.Canceled) {
		warnings = append(warnings, fmt.Sprintf("failed to fetch cover image: %v", err))
	}

	metadata := <-metadataChan

	// Cancellation between the write and the tagging leaves a complete but
	// untagged file. Removing it keeps a cancelled run from being mistaken
	// for a finished one, and nothing has been recorded in the store yet.
	if err := ctx.Err(); err != nil {
		fsutil.Remove(outputPath)
		return downloadResult{err: err}
	}

	warnings = append(warnings, metadata.warnings...)
	warnings = append(warnings, d.finalizeDownload(resource, track, outputPath, outputFormat, metadata.genre, cover, metadata.bpmKey)...)

	return downloadResult{warnings: warnings}
}

// uniqueOutputPath avoids clobbering an unrelated file by appending " (2)",
// " (3)" and so on until the name is free.
//
// The file this track already owns according to the store is exempt: a
// re-download of the same track should overwrite its own output rather than
// pile up numbered copies next to it.
func (d *Downloader) uniqueOutputPath(trackID, path string) string {
	owned := ""
	if info, err := d.store.DownloadInfo(trackID); err == nil {
		owned = info.Path
	}

	ext := filepath.Ext(path)
	stem := strings.TrimSuffix(path, ext)
	candidate := path
	for i := 2; candidate != owned && fsutil.Exists(candidate); i++ {
		candidate = fmt.Sprintf("%s (%d)%s", stem, i, ext)
	}

	return candidate
}

// finalizeDownload tags the finished file and records it in the store,
// returning any non-fatal problems as warnings.
//
// The hash is taken after tagging so it matches the bytes actually on disk,
// which is what the skip check later compares against. The download is
// recorded even when tagging or hashing failed: the audio is there, and
// refusing to record it would mean downloading it all over again next time.
func (d *Downloader) finalizeDownload(resource deezer.Resource, track *deezer.Track, outputPath, outputFormat, genre string, cover []byte, bpmKey bpmKey) []string {
	var warnings []string

	if err := tag.Write(outputPath, buildTagMetadata(resource, track, cover, bpmKey, genre)); err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to add tags: %v", err))
	}

	hash, size, err := hashFile(outputPath)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to get file hash: %v", err))
	}

	info := &store.DownloadInfo{
		TrackID:    track.ID,
		Quality:    outputFormat,
		Path:       outputPath,
		Hash:       hash,
		Size:       size,
		Downloaded: time.Now(),
	}

	if err := d.store.PutDownloadInfo(info); err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to save download info: %v", err))
	}

	return warnings
}
