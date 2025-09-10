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
	"github.com/mathismqn/godeez/internal/crypto"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/fileutil"
	"github.com/mathismqn/godeez/internal/logger"
	"github.com/mathismqn/godeez/internal/provider"
	"github.com/mathismqn/godeez/internal/store"
	"github.com/mathismqn/godeez/internal/tags"
)

const chunkSize = 2048

type Client struct {
	appConfig    *config.Config
	resourceType string
	deezerClient *deezer.Client
	Logger       *logger.Logger

	hashIndexOnce sync.Once
	hashIndex     *fileutil.HashIndex
	hashIndexErr  error
}

func New(appConfig *config.Config, resourceType string) *Client {
	return &Client{
		appConfig:    appConfig,
		resourceType: resourceType,
		deezerClient: nil,
		Logger:       logger.New(nil), // Initialize with a nil logger, can be set later
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

	return c.downloadAllSongs(ctx, resource, id, opts, outputDir)
}

func (c *Client) initDeezerClient(ctx context.Context, opts Options) error {
	var err error
	c.deezerClient, err = deezer.NewClient(ctx, c.appConfig)
	if err != nil {
		return err
	}

	if !c.deezerClient.Session.Premium && (opts.Quality == "mp3_320" || opts.Quality == "flac") {
		return fmt.Errorf("premium account required for '%s' quality", opts.Quality)
	}

	return nil
}

func (c *Client) prepareResource(ctx context.Context, id string, opts Options) (deezer.Resource, string, error) {
	resource, err := c.createResource()
	if err != nil {
		return nil, "", err
	}

	if err := c.deezerClient.FetchResource(ctx, resource, id); err != nil {
		return nil, "", fmt.Errorf("failed to fetch resource: %w", err)
	}

	songs := resource.GetSongs()
	if len(songs) == 0 {
		if c.resourceType == "track" {
			return nil, "", fmt.Errorf("track with ID %s not found", id)
		}
		return nil, "", fmt.Errorf("%s has no songs", c.resourceType)
	}

	if c.resourceType == "artist" && len(songs) > opts.Limit {
		songs = songs[:opts.Limit]
		resource.SetSongs(songs)
	}

	resourceOutputDir := resource.GetOutputDir(c.appConfig.OutputDir)
	if err := fileutil.EnsureDir(resourceOutputDir); err != nil {
		return nil, "", fmt.Errorf("failed to create output directory: %w", err)
	}

	return resource, resourceOutputDir, nil
}

func (c *Client) createResource() (deezer.Resource, error) {
	switch c.resourceType {
	case "album":
		return &deezer.Album{}, nil
	case "playlist":
		return &deezer.Playlist{}, nil
	case "artist":
		return &deezer.Artist{}, nil
	case "track":
		return &deezer.Track{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", c.resourceType)
	}
}

func (c *Client) downloadAllSongs(ctx context.Context, resource deezer.Resource, resourceID string, opts Options, outputDir string) error {
	songs := resource.GetSongs()
	startTime := time.Now()

	if c.resourceType != "track" {
		fmt.Printf("%s\n\nStarting download...\n\n", resource)
	}

	progress := newProgressTracker(c.Logger, len(songs), c.resourceType)

	for i, song := range songs {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		sp := progress.startDownload(i, song)
		result := c.downloadSong(ctx, resource, song, opts, outputDir)
		sp.Stop()

		if result.err != nil {
			if errors.Is(result.err, context.Canceled) {
				return result.err
			}
		}

		progress.handleResult(i, song, result)
	}

	progress.printSummary(resource.GetTitle(), resourceID, outputDir, time.Since(startTime))

	return nil
}

func (c *Client) downloadSong(ctx context.Context, resource deezer.Resource, song *deezer.Song, opts Options, outputDir string) downloadResult {
	var warnings []string

	media, err := c.deezerClient.FetchMedia(ctx, song, opts.Quality)
	if err != nil {
		return handleError(fmt.Errorf("failed to fetch media: %w", err))
	}

	mediaFormat := media.GetFormat()
	if opts.Strict && strings.ToLower(mediaFormat) != opts.Quality {
		return handleError(fmt.Errorf("requested quality '%s' not available", opts.Quality))
	}

	if path, skip := c.shouldSkipDownload(ctx, song.ID, mediaFormat); skip {
		return handleError(SkipError{Path: path})
	}

	metadataFetcher := newMetadataFetcher(c.deezerClient.Session.HttpClient)
	metadataResult := metadataFetcher.fetch(ctx, song, opts)
	warnings = append(warnings, metadataResult.warnings...)

	stream, err := c.deezerClient.GetMediaStream(ctx, media, song.ID)
	if err != nil {
		return handleError(fmt.Errorf("failed to get media stream: %w", err))
	}

	dlCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	fileName := song.GetFileName(c.resourceType, mediaFormat, song)
	outputPath := path.Join(outputDir, fileName)

	key := crypto.GetKey(c.appConfig.SecretKey, song.ID)
	if err := c.streamToFile(dlCtx, stream, outputPath, key); err != nil {
		fileutil.DeleteFile(outputPath)
		return handleError(fmt.Errorf("failed to stream to file: %w", err))
	}

	if opts.Quality != strings.ToLower(mediaFormat) {
		warnings = append(warnings, fmt.Sprintf("requested quality '%s' not available, using '%s' instead", opts.Quality, strings.ToLower(mediaFormat)))
	}

	cover, err := c.deezerClient.FetchCoverImage(ctx, song)
	if err != nil && !errors.Is(err, context.Canceled) {
		warnings = append(warnings, fmt.Sprintf("failed to fetch cover image: %v", err))
	}

	finalizeWarnings := c.finalizeDownload(resource, song, outputPath, mediaFormat, metadataResult.genre, cover, metadataResult.bpmKey)
	warnings = append(warnings, finalizeWarnings...)

	return downloadResult{
		success:  true,
		warnings: warnings,
	}
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
			// continue
		}

		totalRead := 0
		for totalRead < chunkSize {
			n, err := stream.Read(buffer[totalRead:])
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			if n > 0 {
				totalRead += n
			}
		}

		if totalRead == 0 {
			break
		}

		if chunk%3 == 0 && totalRead == chunkSize {
			buffer, err = crypto.Decrypt(buffer, key)
			if err != nil {
				return err
			}
		}

		_, err = file.Write(buffer[:totalRead])
		if err != nil {
			return err
		}

		if totalRead < chunkSize {
			break
		}
	}

	return nil
}

func (c *Client) finalizeDownload(resource deezer.Resource, song *deezer.Song, outputPath, mediaFormat, genre string, cover []byte, bpmKey provider.BPMKey) []string {
	var warnings []string

	if err := tags.AddTags(resource, song, cover, outputPath, bpmKey.BPM, bpmKey.Key, genre); err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to add tags: %v", err))
	}

	hash, err := fileutil.GetFileHash(outputPath)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to get file hash: %v", err))
	}

	info := &store.DownloadInfo{
		SongID:     song.ID,
		Quality:    mediaFormat,
		Path:       outputPath,
		Hash:       hash,
		Downloaded: time.Now(),
	}

	if err := info.Save(); err != nil {
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
