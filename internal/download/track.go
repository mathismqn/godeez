package download

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/fsutil"
	"github.com/mathismqn/godeez/internal/store"
	"github.com/mathismqn/godeez/internal/tag"
)

func (d *Downloader) downloadTrack(ctx context.Context, resource deezer.Resource, track *deezer.Track, opts Options, outputDir string) downloadResult {
	media, err := d.deezerClient.FetchMedia(ctx, track, opts.Quality)
	if err != nil {
		return downloadResult{err: fmt.Errorf("failed to fetch media: %w", err)}
	}

	mediaFormat := media.GetFormat()
	if opts.Strict && strings.ToLower(mediaFormat) != opts.Quality {
		return downloadResult{err: fmt.Errorf("requested quality '%s' not available", opts.Quality)}
	}

	if skipPath, skip := d.shouldSkipDownload(ctx, track.ID, mediaFormat); skip {
		return downloadResult{skipped: true, path: skipPath}
	}

	metadataChan := make(chan metadataResult, 1)
	go func() {
		metadataChan <- fetchMetadata(d.deezerClient.Session.HttpClient, ctx, track, opts)
	}()

	stream, err := d.deezerClient.GetMediaStream(ctx, media)
	if err != nil {
		return downloadResult{err: fmt.Errorf("failed to get media stream: %w", err)}
	}

	dlCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	fileName := track.Filename(d.kind, mediaFormat)
	outputPath := path.Join(outputDir, fileName)

	key := deezer.BlowfishKey(track.ID)
	if err := d.streamToFile(dlCtx, stream, outputPath, key); err != nil {
		fsutil.Remove(outputPath)
		return downloadResult{err: fmt.Errorf("failed to stream to file: %w", err)}
	}

	var warnings []string

	if opts.Quality != strings.ToLower(mediaFormat) {
		warnings = append(warnings, fmt.Sprintf("requested quality '%s' not available, using '%s' instead", opts.Quality, strings.ToLower(mediaFormat)))
	}

	cover, err := d.deezerClient.FetchCoverImage(ctx, track)
	if err != nil && !errors.Is(err, context.Canceled) {
		warnings = append(warnings, fmt.Sprintf("failed to fetch cover image: %v", err))
	}

	metadata := <-metadataChan
	warnings = append(warnings, metadata.warnings...)
	warnings = append(warnings, d.finalizeDownload(resource, track, outputPath, mediaFormat, metadata.genre, cover, metadata.bpmKey)...)

	return downloadResult{warnings: warnings}
}

func (d *Downloader) finalizeDownload(resource deezer.Resource, track *deezer.Track, outputPath, mediaFormat, genre string, cover []byte, bpmKey bpmKey) []string {
	var warnings []string

	if err := tag.Write(outputPath, buildTagMetadata(resource, track, cover, bpmKey, genre)); err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to add tags: %v", err))
	}

	hash, err := hashFile(outputPath)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to get file hash: %v", err))
	}

	info := &store.DownloadInfo{
		TrackID:    track.ID,
		Quality:    mediaFormat,
		Path:       outputPath,
		Hash:       hash,
		Downloaded: time.Now(),
	}

	if err := d.store.PutDownloadInfo(info); err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to save download info: %v", err))
	}

	return warnings
}
