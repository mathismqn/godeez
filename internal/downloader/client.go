package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/mathismqn/godeez/internal/app"
	"github.com/mathismqn/godeez/internal/bpm"
	"github.com/mathismqn/godeez/internal/crypto"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/fileutil"
	"github.com/mathismqn/godeez/internal/store"
	"github.com/mathismqn/godeez/internal/tags"
)

const chunkSize = 2048

type Client struct {
	appCtx       *app.Context
	resourceType string
	deezerClient *deezer.Client

	hashIndexOnce sync.Once
	hashIndex     *fileutil.HashIndex
	hashIndexErr  error
}

func New(appCtx *app.Context, resourceType string) *Client {
	return &Client{
		appCtx:       appCtx,
		resourceType: resourceType,
		deezerClient: nil,
	}
}

func (c *Client) Run(ctx context.Context, opts Options, ids []string) error {
	var err error
	c.deezerClient, err = deezer.NewClient(ctx, c.appCtx)
	if err != nil {
		return err
	}

	for _, id := range ids {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var resource deezer.Resource

		switch c.resourceType {
		case "album":
			resource = &deezer.Album{}
		case "playlist":
			resource = &deezer.Playlist{}
		default:
			return fmt.Errorf("unsupported resource type: %s", c.resourceType)
		}

		if err := c.deezerClient.FetchResource(ctx, resource, id); err != nil {
			return fmt.Errorf("failed to fetch resource: %w", err)
		}

		songs := resource.GetSongs()
		if len(songs) == 0 {
			return fmt.Errorf("%s has no songs", c.resourceType)
		}

		outputDir := resource.GetOutputDir(opts.OutputDir)
		if err := fileutil.EnsureDir(outputDir); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		for _, song := range songs {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err := c.downloadSong(ctx, resource, song, opts, outputDir); err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}

				fmt.Fprintf(os.Stderr, "Error: failed to download %s: %v\n", song.Title, err)
				continue
			}
		}
	}

	return nil
}

func (c *Client) downloadSong(ctx context.Context, resource deezer.Resource, song *deezer.Song, opts Options, outputDir string) error {
	media, err := c.deezerClient.FetchMedia(ctx, song, opts.Quality)
	if err != nil {
		return err
	}

	fileName := song.GetFileName(c.resourceType, song, media)
	outputPath := path.Join(outputDir, fileName)

	mediaFormat, err := media.GetFormat()
	if err != nil {
		return err
	}

	if _, skip := c.shouldSkipDownload(ctx, song.ID, mediaFormat); skip {
		return nil
	}

	metricsChan := make(chan *bpm.Metrics, 1)
	errChan := make(chan error, 1)

	go func() {
		metrics, err := bpm.FetchMetrics(ctx, c.deezerClient.Session.HttpClient, song.Artist, song.GetTitle(), song.Duration)
		if err != nil {
			errChan <- err
			return
		}

		metricsChan <- metrics
	}()

	stream, err := c.deezerClient.GetMediaStream(ctx, media, song.ID)
	if err != nil {
		return fmt.Errorf("media stream unavailable: %w", err)
	}

	dlCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	key := crypto.GetKey(c.appCtx.Config.SecretKey, song.ID)
	if err := c.streamToFile(dlCtx, stream, outputPath, key); err != nil {
		fileutil.DeleteFile(outputPath)

		return fmt.Errorf("unable to write to file: %w", err)
	}

	metrics := &bpm.Metrics{}
	select {
	case metrics = <-metricsChan:
		fmt.Printf("BPM: %s, Key: %s\n", metrics.BPM, metrics.Key)
	case err := <-errChan:
		if !errors.Is(err, context.Canceled) {
			fmt.Printf("Warning: failed to fetch BPM and key: %v\n", err)
		}
	}

	cover, err := c.deezerClient.FetchCoverImage(ctx, song)
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("Warning: failed to fetch cover image: %v\n", err)
	}

	c.finalizeDownload(resource, song, outputPath, mediaFormat, cover, metrics)

	return nil
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

func (c *Client) finalizeDownload(resource deezer.Resource, song *deezer.Song, outputPath, mediaFormat string, cover []byte, metrics *bpm.Metrics) {
	if err := tags.AddTags(resource, song, cover, outputPath, metrics.BPM, metrics.Key); err != nil {
		fmt.Printf("Warning: failed to add tags: %v\n", err)
	}

	hash, err := fileutil.GetFileHash(outputPath)
	if err != nil {
		fmt.Printf("Warning: failed to get file hash: %v\n", err)
	}

	info := &store.DownloadInfo{
		SongID:     song.ID,
		Quality:    mediaFormat,
		Path:       outputPath,
		Hash:       hash,
		Downloaded: time.Now(),
	}

	if err := info.Save(); err != nil {
		fmt.Printf("Warning: failed to save download info: %v\n", err)
	}
}

func (c *Client) initHashIndex(ctx context.Context) error {
	c.hashIndexOnce.Do(func() {
		c.hashIndex, c.hashIndexErr = fileutil.NewHashIndex(ctx, c.appCtx.AppDir)
	})

	return c.hashIndexErr
}
