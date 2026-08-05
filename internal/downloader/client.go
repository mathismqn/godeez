package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/fileutil"
	"github.com/mathismqn/godeez/internal/store"
	"github.com/mathismqn/godeez/internal/tag"
)

const chunkSize = 2048

type Client struct {
	appConfig    *config.Config
	store        *store.Store
	kind         deezer.Kind
	deezerClient *deezer.Client

	hashIndexOnce sync.Once
	hashIndex     *fileutil.HashIndex
	hashIndexErr  error
}

func New(appConfig *config.Config, st *store.Store, kind deezer.Kind) *Client {
	return &Client{
		appConfig: appConfig,
		store:     st,
		kind:      kind,
	}
}

func (c *Client) Run(ctx context.Context, opts Options, id string) error {
	if err := c.initDeezerClient(ctx, opts); err != nil {
		return err
	}

	resource, outputDir, err := c.prepareResource(ctx, id, opts)
	if err != nil {
		return err
	}

	return c.downloadAllTracks(ctx, resource, opts, outputDir)
}

func (c *Client) initDeezerClient(ctx context.Context, opts Options) error {
	var err error
	c.deezerClient, err = deezer.NewClient(ctx, c.appConfig.ARLCookie)
	if err != nil {
		return err
	}

	if !c.deezerClient.Session.Premium && (opts.Quality == "mp3_320" || opts.Quality == "flac") {
		return fmt.Errorf("premium account required for '%s' quality", opts.Quality)
	}

	return nil
}

func (c *Client) prepareResource(ctx context.Context, id string, opts Options) (deezer.Resource, string, error) {
	resource, err := c.deezerClient.FetchResource(ctx, c.kind, id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch resource: %w", err)
	}

	tracks := resource.GetTracks()
	if len(tracks) == 0 {
		if c.kind == deezer.KindTrack {
			return nil, "", fmt.Errorf("track with ID %s not found", id)
		}
		return nil, "", fmt.Errorf("%s has no tracks", c.kind)
	}

	if c.kind == deezer.KindArtist && len(tracks) > opts.Limit {
		resource.SetTracks(tracks[:opts.Limit])
	}

	outputDir := resource.GetOutputDir(c.appConfig.OutputDir)
	if err := fileutil.EnsureDir(outputDir); err != nil {
		return nil, "", fmt.Errorf("failed to create output directory: %w", err)
	}

	return resource, outputDir, nil
}

func (c *Client) downloadAllTracks(ctx context.Context, resource deezer.Resource, opts Options, outputDir string) error {
	tracks := resource.GetTracks()
	startTime := time.Now()

	if c.kind != deezer.KindTrack {
		fmt.Printf("%s\n\nStarting download...\n\n", resourceInfo(resource))
	}

	progress := newProgressTracker(len(tracks), c.kind)

	for i, track := range tracks {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		sp := progress.startDownload(i, track)
		result := c.downloadTrack(ctx, resource, track, opts, outputDir)
		sp.Stop()

		if result.err != nil && errors.Is(result.err, context.Canceled) {
			return result.err
		}

		progress.handleResult(i, track, result)
	}

	progress.printSummary(outputDir, time.Since(startTime))

	return nil
}

func (c *Client) downloadTrack(ctx context.Context, resource deezer.Resource, track *deezer.Track, opts Options, outputDir string) downloadResult {
	media, err := c.deezerClient.FetchMedia(ctx, track, opts.Quality)
	if err != nil {
		return downloadResult{err: fmt.Errorf("failed to fetch media: %w", err)}
	}

	mediaFormat := media.GetFormat()
	if opts.Strict && strings.ToLower(mediaFormat) != opts.Quality {
		return downloadResult{err: fmt.Errorf("requested quality '%s' not available", opts.Quality)}
	}

	if skipPath, skip := c.shouldSkipDownload(ctx, track.ID, mediaFormat); skip {
		return downloadResult{skipped: true, path: skipPath}
	}

	metadataChan := make(chan metadataResult, 1)
	go func() {
		metadataChan <- fetchMetadata(c.deezerClient.Session.HttpClient, ctx, track, opts)
	}()

	stream, err := c.deezerClient.GetMediaStream(ctx, media)
	if err != nil {
		return downloadResult{err: fmt.Errorf("failed to get media stream: %w", err)}
	}

	dlCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	fileName := track.Filename(c.kind, mediaFormat)
	outputPath := path.Join(outputDir, fileName)

	key := deezer.BlowfishKey(track.ID)
	if err := c.streamToFile(dlCtx, stream, outputPath, key); err != nil {
		fileutil.DeleteFile(outputPath)
		return downloadResult{err: fmt.Errorf("failed to stream to file: %w", err)}
	}

	var warnings []string

	if opts.Quality != strings.ToLower(mediaFormat) {
		warnings = append(warnings, fmt.Sprintf("requested quality '%s' not available, using '%s' instead", opts.Quality, strings.ToLower(mediaFormat)))
	}

	cover, err := c.deezerClient.FetchCoverImage(ctx, track)
	if err != nil && !errors.Is(err, context.Canceled) {
		warnings = append(warnings, fmt.Sprintf("failed to fetch cover image: %v", err))
	}

	metadata := <-metadataChan
	warnings = append(warnings, metadata.warnings...)
	warnings = append(warnings, c.finalizeDownload(resource, track, outputPath, mediaFormat, metadata.genre, cover, metadata.bpmKey)...)

	return downloadResult{warnings: warnings}
}

func (c *Client) streamToFile(ctx context.Context, stream io.ReadCloser, outputPath string, key []byte) error {
	defer stream.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, chunkSize)
	for chunk := 0; ; chunk++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		totalRead := 0
		for totalRead < chunkSize {
			n, err := stream.Read(buffer[totalRead:])
			totalRead += n
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
		}

		if totalRead == 0 {
			break
		}

		if chunk%3 == 0 && totalRead == chunkSize {
			buffer, err = deezer.DecryptBlowfish(buffer, key)
			if err != nil {
				return err
			}
		}

		if _, err = file.Write(buffer[:totalRead]); err != nil {
			return err
		}

		if totalRead < chunkSize {
			break
		}
	}

	return nil
}

func (c *Client) finalizeDownload(resource deezer.Resource, track *deezer.Track, outputPath, mediaFormat, genre string, cover []byte, bpmKey bpmKey) []string {
	var warnings []string

	if err := tag.Write(outputPath, buildTagMetadata(resource, track, cover, bpmKey, genre)); err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to add tags: %v", err))
	}

	hash, err := fileutil.GetFileHash(outputPath)
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

	if err := c.store.PutDownloadInfo(info); err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to save download info: %v", err))
	}

	return warnings
}

func (c *Client) initHashIndex(ctx context.Context) error {
	c.hashIndexOnce.Do(func() {
		c.hashIndex, c.hashIndexErr = fileutil.NewHashIndex(ctx, c.appConfig.OutputDir)
	})

	return c.hashIndexErr
}
