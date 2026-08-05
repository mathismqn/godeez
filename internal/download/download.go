package download

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/fsutil"
	"github.com/mathismqn/godeez/internal/store"
)

type Downloader struct {
	appConfig    *config.Config
	store        *store.Store
	kind         deezer.Kind
	deezerClient *deezer.Client

	hashIndexOnce sync.Once
	hashIndex     *hashIndex
	hashIndexErr  error
}

func New(appConfig *config.Config, st *store.Store, kind deezer.Kind) *Downloader {
	return &Downloader{
		appConfig: appConfig,
		store:     st,
		kind:      kind,
	}
}

func (d *Downloader) Run(ctx context.Context, opts Options, id string) error {
	if err := d.initDeezerClient(ctx, opts); err != nil {
		return err
	}

	resource, outputDir, err := d.prepareResource(ctx, id, opts)
	if err != nil {
		return err
	}

	return d.downloadAllTracks(ctx, resource, opts, outputDir)
}

func (d *Downloader) initDeezerClient(ctx context.Context, opts Options) error {
	var err error
	d.deezerClient, err = deezer.NewClient(ctx, d.appConfig.ARLCookie)
	if err != nil {
		return err
	}

	if !d.deezerClient.Session.Premium && (opts.Quality == "mp3_320" || opts.Quality == "flac") {
		return fmt.Errorf("premium account required for '%s' quality", opts.Quality)
	}

	return nil
}

func (d *Downloader) prepareResource(ctx context.Context, id string, opts Options) (deezer.Resource, string, error) {
	resource, err := d.deezerClient.FetchResource(ctx, d.kind, id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch resource: %w", err)
	}

	tracks := resource.Tracks()
	if len(tracks) == 0 {
		if d.kind == deezer.KindTrack {
			return nil, "", fmt.Errorf("track with ID %s not found", id)
		}
		return nil, "", fmt.Errorf("%s has no tracks", d.kind)
	}

	if d.kind == deezer.KindArtist && len(tracks) > opts.Limit {
		resource.SetTracks(tracks[:opts.Limit])
	}

	outputDir := resource.OutputDir(d.appConfig.OutputDir)
	if err := fsutil.EnsureDir(outputDir); err != nil {
		return nil, "", fmt.Errorf("failed to create output directory: %w", err)
	}
	sweepPartFiles(outputDir)

	return resource, outputDir, nil
}

func (d *Downloader) downloadAllTracks(ctx context.Context, resource deezer.Resource, opts Options, outputDir string) error {
	tracks := resource.Tracks()
	startTime := time.Now()

	if d.kind != deezer.KindTrack {
		fmt.Printf("%s\n\nStarting download...\n\n", resourceInfo(resource))
	}

	progress := newProgressTracker(len(tracks), d.kind)

	for i, track := range tracks {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		sp := progress.startDownload(i, track)
		result := d.downloadTrack(ctx, resource, track, opts, outputDir)
		sp.Stop()

		if result.err != nil && errors.Is(result.err, context.Canceled) {
			return result.err
		}

		progress.handleResult(i, track, result)
	}

	progress.printSummary(outputDir, time.Since(startTime))

	return nil
}
